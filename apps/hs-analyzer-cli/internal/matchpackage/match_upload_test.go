package matchpackage

import (
	"testing"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/entity"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/monitor"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
	"hsanalyzer/pkg/contracts/matchupload"
)

func TestBuildMatchUploadPayloadAfterGameEnd(t *testing.T) {
	t.Parallel()

	gameState := state.New("game-1", state.GameModeStandard)
	match := monitor.PersistedMatch{
		Key: "game:game-1",
		Events: []game.Event{
			game.NewGameStartEvent("game-1", state.GameModeStandard),
			game.NewTurnStartEvent(types.PlayerTypePlayer),
			game.NewGameEndEvent(types.PlayerTypePlayer),
		},
		State:    gameState,
		Complete: true,
		Reason:   monitor.CompletionReasonGameEnd,
	}

	payload := BuildMatchUploadPayload(match, "test")
	if payload.SchemaVersion != matchupload.SchemaVersionV1 {
		t.Fatalf("unexpected schema version: %s", payload.SchemaVersion)
	}
	if payload.MatchID != "game:game-1" {
		t.Fatalf("unexpected match id: %s", payload.MatchID)
	}
	if payload.GameID != "game-1" {
		t.Fatalf("unexpected game id: %s", payload.GameID)
	}
	if payload.CompletionStatus != matchupload.CompletionStatusComplete {
		t.Fatalf("unexpected completion status: %s", payload.CompletionStatus)
	}
	if payload.CompletionReason != matchupload.CompletionReasonGameEnd {
		t.Fatalf("unexpected completion reason: %s", payload.CompletionReason)
	}
	if len(payload.Events) != 3 {
		t.Fatalf("unexpected events count: %d", len(payload.Events))
	}
	_, ok, err := payload.GetDashboardSummary()
	if err != nil {
		t.Fatalf("decode dashboard summary: %v", err)
	}
	if ok {
		t.Fatalf("did not expect dashboard summary without deck list event")
	}
	if err := payload.Validate(); err != nil {
		t.Fatalf("validate payload: %v", err)
	}
}

func TestBuildMatchUploadPayloadForIncompleteFallback(t *testing.T) {
	t.Parallel()

	match := monitor.PersistedMatch{
		Key: "events:fallback",
		Events: []game.Event{
			game.NewGameStartEvent("", state.GameModeStandard),
			game.NewTurnStartEvent(types.PlayerTypeOpponent),
		},
		State:    state.New("", state.GameModeStandard),
		Complete: false,
		Reason:   monitor.CompletionReasonNextGameStart,
	}

	payload := BuildMatchUploadPayload(match, "")
	if payload.CompletionStatus != matchupload.CompletionStatusIncomplete {
		t.Fatalf("unexpected completion status: %s", payload.CompletionStatus)
	}
	if payload.CompletionReason != matchupload.CompletionReasonNextGameStart {
		t.Fatalf("unexpected completion reason: %s", payload.CompletionReason)
	}
	if err := payload.Validate(); err != nil {
		t.Fatalf("validate incomplete payload: %v", err)
	}
}

func TestBuildMatchUploadPayloadAddsDashboardSummaryFromDeckList(t *testing.T) {
	t.Parallel()

	gameState := state.New("game-2", state.GameModeStandard)
	winner := types.PlayerTypePlayer
	gameState.Winner = &winner
	gameState.Opponent.Hero = entity.New(900, "HERO_01", types.PlayerTypeOpponent)

	first := card.New("CARD_001", 1, types.PlayerTypePlayer)
	first.Info = &carddb.CardInfo{
		ID:     "CARD_001",
		Name:   "Taste of Chaos",
		Cost:   1,
		Rarity: carddb.RarityCommon,
		Class:  carddb.CardClassDemonHunter,
	}
	second := card.New("CARD_001", 2, types.PlayerTypePlayer)
	second.Info = first.Info
	legendary := card.New("CARD_999", 3, types.PlayerTypePlayer)
	legendary.Info = &carddb.CardInfo{
		ID:     "CARD_999",
		Name:   "Argus, the Emerald Star",
		Cost:   5,
		Rarity: carddb.RarityLegendary,
		Class:  carddb.CardClassDemonHunter,
	}

	match := monitor.PersistedMatch{
		Key: "game:game-2",
		Events: []game.Event{
			game.NewDeckListEvent(types.PlayerTypePlayer, []*card.Card{first, second, legendary}, "Spell Demon Hunter", "deck-spell-dh", "code"),
			game.NewGameStartEvent("game-2", state.GameModeStandard),
			game.NewGameEndEvent(types.PlayerTypePlayer),
		},
		State:    gameState,
		Complete: true,
		Reason:   monitor.CompletionReasonGameEnd,
	}

	payload := BuildMatchUploadPayload(match, "test")
	summary, ok, err := payload.GetDashboardSummary()
	if err != nil {
		t.Fatalf("decode dashboard summary: %v", err)
	}
	if !ok {
		t.Fatalf("expected dashboard summary")
	}
	if summary.DeckID != "deck-spell-dh" {
		t.Fatalf("unexpected deck id: %s", summary.DeckID)
	}
	if summary.DeckName != "Spell Demon Hunter" {
		t.Fatalf("unexpected deck name: %s", summary.DeckName)
	}
	if summary.PlayerClass != "DEMONHUNTER" {
		t.Fatalf("unexpected player class: %s", summary.PlayerClass)
	}
	if summary.OpponentClass != "WARRIOR" {
		t.Fatalf("unexpected opponent class: %s", summary.OpponentClass)
	}
	if summary.Result != matchupload.DashboardResultWin {
		t.Fatalf("unexpected result: %s", summary.Result)
	}
	if len(summary.DeckCards) != 2 {
		t.Fatalf("unexpected deck card count: %d", len(summary.DeckCards))
	}
	if summary.DeckCards[0].Copies != 2 {
		t.Fatalf("unexpected copies for first card: %d", summary.DeckCards[0].Copies)
	}
	if !summary.DeckCards[1].IsLegendary {
		t.Fatalf("expected second card to be legendary")
	}
}
