package monitor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
	"hsanalyzer/pkg/contracts/overlay"
)

type recordingOverlaySink struct {
	snapshots []overlay.Snapshot
}

func (s *recordingOverlaySink) Publish(snapshot overlay.Snapshot) error {
	s.snapshots = append(s.snapshots, snapshot)
	return nil
}

func TestOverlaySnapshotTracksDeckRemainingAndOpponentCardsSeparately(t *testing.T) {
	t.Parallel()

	projector := NewOverlaySnapshotProjector(OverlaySnapshotProjectorOptions{
		CardDB: &stubLiveCardStatsCardDB{
			cards: map[string]*carddb.CardInfo{
				"DECK_1": {ID: "DECK_1", Name: "Моя карта", Cost: 2, Text: "[x]Описание", Type: carddb.CardTypeSpell, Rarity: carddb.RarityEpic},
				"OPP_1":  {ID: "OPP_1", Name: "Карта оппонента", Cost: 3},
			},
		},
		Now: fixedTimeSequence(time.Unix(100, 0), time.Unix(101, 0)),
	})

	mustOverlayProject(t, projector, game.NewDeckListEvent(types.PlayerTypePlayer, []*card.Card{
		card.New("DECK_1", 1, types.PlayerTypePlayer),
		card.New("DECK_1", 2, types.PlayerTypePlayer),
	}, "Spell DH", "deck-1", ""))
	mustOverlayProject(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustOverlayProject(t, projector, game.NewCardDrawnEvent(types.PlayerTypePlayer, card.New("DECK_1", 3, types.PlayerTypePlayer)))
	mustOverlayProject(t, projector, game.NewCardPlayedEvent(types.PlayerTypePlayer, card.New("DECK_1", 4, types.PlayerTypePlayer), 0))
	mustOverlayProject(t, projector, game.NewCardPlayedEvent(types.PlayerTypeOpponent, card.New("OPP_1", 4, types.PlayerTypeOpponent), 0))

	snapshot := projector.Snapshot()
	if snapshot.SchemaVersion != overlay.SchemaVersionV1 {
		t.Fatalf("unexpected schema version: %q", snapshot.SchemaVersion)
	}
	if len(snapshot.PlayerDeckRemaining) != 1 {
		t.Fatalf("unexpected player deck entries: %#v", snapshot.PlayerDeckRemaining)
	}
	if len(snapshot.OpponentPlayedCards) != 1 {
		t.Fatalf("unexpected opponent entries: %#v", snapshot.OpponentPlayedCards)
	}
	if snapshot.PlayerDeckRemaining[0].RemainingCopies != 1 || snapshot.PlayerDeckRemaining[0].DrawnCopies != 1 || snapshot.PlayerDeckRemaining[0].PlayedCopies != 1 {
		t.Fatalf("unexpected player deck counters: %#v", snapshot.PlayerDeckRemaining[0])
	}
	if snapshot.PlayerDeckRemaining[0].Rarity != string(carddb.RarityEpic) || snapshot.PlayerDeckRemaining[0].CardType != string(carddb.CardTypeSpell) {
		t.Fatalf("unexpected player deck metadata: %#v", snapshot.PlayerDeckRemaining[0])
	}
	if snapshot.OpponentPlayedCards[0].PlayCount != 1 {
		t.Fatalf("unexpected opponent play count: %#v", snapshot.OpponentPlayedCards[0])
	}
}

func TestOverlaySnapshotIgnoresGeneratedCardsForRemainingDeck(t *testing.T) {
	t.Parallel()

	projector := NewOverlaySnapshotProjector(OverlaySnapshotProjectorOptions{})
	mustOverlayProject(t, projector, game.NewDeckListEvent(types.PlayerTypePlayer, []*card.Card{
		card.New("DECK_1", 1, types.PlayerTypePlayer),
	}, "Deck", "deck-1", ""))
	mustOverlayProject(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustOverlayProject(t, projector, game.NewCardDrawnEvent(types.PlayerTypePlayer, card.New("GENERATED_1", 2, types.PlayerTypePlayer)))

	snapshot := projector.Snapshot()
	if snapshot.PlayerDeckRemaining[0].RemainingCopies != 1 || snapshot.PlayerDeckRemaining[0].DrawnCopies != 0 {
		t.Fatalf("unexpected deck counters after generated card draw: %#v", snapshot.PlayerDeckRemaining[0])
	}
}

func TestOverlaySnapshotReplaysPreGameOpeningHandOnGameStart(t *testing.T) {
	t.Parallel()

	projector := NewOverlaySnapshotProjector(OverlaySnapshotProjectorOptions{})
	mustOverlayProject(t, projector, game.NewDeckListEvent(types.PlayerTypePlayer, []*card.Card{
		card.New("OPENING_1", 1, types.PlayerTypePlayer),
		card.New("OPENING_1", 2, types.PlayerTypePlayer),
		card.New("OPENING_2", 3, types.PlayerTypePlayer),
	}, "Deck", "deck-1", ""))

	mustOverlayProject(t, projector, game.NewCardDrawnEvent(types.PlayerTypePlayer, card.New("OPENING_1", 10, types.PlayerTypePlayer)))
	mustOverlayProject(t, projector, game.NewCardDrawnEvent(types.PlayerTypeOpponent, card.New("OPENING_2", 11, types.PlayerTypeOpponent)))
	mustOverlayProject(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))

	opening1 := requireOverlayDeckCard(t, projector.Snapshot(), "OPENING_1")
	if opening1.RemainingCopies != 1 || opening1.DrawnCopies != 1 || opening1.PlayedCopies != 0 {
		t.Fatalf("unexpected first opening card counters: %#v", opening1)
	}

	opening2 := requireOverlayDeckCard(t, projector.Snapshot(), "OPENING_2")
	if opening2.RemainingCopies != 0 || opening2.DrawnCopies != 1 || opening2.PlayedCopies != 0 {
		t.Fatalf("expected raw-side opening draw to be counted in hand: %#v", opening2)
	}

	mustOverlayProject(t, projector, game.NewCardPlayedEvent(types.PlayerTypePlayer, card.New("OPENING_2", 11, types.PlayerTypePlayer), 0))
	opening2 = requireOverlayDeckCard(t, projector.Snapshot(), "OPENING_2")
	if opening2.RemainingCopies != 0 || opening2.DrawnCopies != 1 || opening2.PlayedCopies != 1 {
		t.Fatalf("expected played opening card to move from hand to other: %#v", opening2)
	}
}

func TestOverlaySnapshotReplaysPreGameOpponentPlaysOnGameStart(t *testing.T) {
	t.Parallel()

	projector := NewOverlaySnapshotProjector(OverlaySnapshotProjectorOptions{
		Now: fixedTimeSequence(time.Unix(100, 0), time.Unix(101, 0)),
	})

	mustOverlayProject(t, projector, game.NewCardPlayedEvent(types.PlayerTypeOpponent, card.New("OPP_EARLY", 20, types.PlayerTypeOpponent), 0))
	mustOverlayProject(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))

	snapshot := projector.Snapshot()
	if len(snapshot.OpponentPlayedCards) != 1 {
		t.Fatalf("expected early opponent play to be replayed after game start: %#v", snapshot.OpponentPlayedCards)
	}
	if snapshot.OpponentPlayedCards[0].CardID != "OPP_EARLY" || snapshot.OpponentPlayedCards[0].PlayCount != 1 {
		t.Fatalf("unexpected early opponent play: %#v", snapshot.OpponentPlayedCards[0])
	}
}

func TestOverlaySnapshotBackfillsRemainingDeckWhenSelfPlayArrivesWithoutDrawEvent(t *testing.T) {
	t.Parallel()

	projector := NewOverlaySnapshotProjector(OverlaySnapshotProjectorOptions{})
	mustOverlayProject(t, projector, game.NewDeckListEvent(types.PlayerTypePlayer, []*card.Card{
		card.New("DECK_1", 1, types.PlayerTypePlayer),
		card.New("DECK_1", 2, types.PlayerTypePlayer),
	}, "Deck", "deck-1", ""))
	mustOverlayProject(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustOverlayProject(t, projector, game.NewCardPlayedEvent(types.PlayerTypePlayer, card.New("DECK_1", 3, types.PlayerTypePlayer), 0))

	snapshot := projector.Snapshot()
	if snapshot.PlayerDeckRemaining[0].RemainingCopies != 1 || snapshot.PlayerDeckRemaining[0].DrawnCopies != 1 || snapshot.PlayerDeckRemaining[0].PlayedCopies != 1 {
		t.Fatalf("unexpected deck counters after self play fallback: %#v", snapshot.PlayerDeckRemaining[0])
	}
}

func TestOverlaySnapshotAggregatesOpponentRepeatsAndResetsOnNewMatch(t *testing.T) {
	t.Parallel()

	projector := NewOverlaySnapshotProjector(OverlaySnapshotProjectorOptions{
		Now: fixedTimeSequence(
			time.Unix(100, 0),
			time.Unix(101, 0),
			time.Unix(102, 0),
			time.Unix(103, 0),
		),
	})

	mustOverlayProject(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustOverlayProject(t, projector, game.NewCardPlayedEvent(types.PlayerTypeOpponent, card.New("OPP_1", 1, types.PlayerTypeOpponent), 0))
	mustOverlayProject(t, projector, game.NewCardPlayedEvent(types.PlayerTypeOpponent, card.New("OPP_1", 2, types.PlayerTypeOpponent), 0))

	snapshot := projector.Snapshot()
	if snapshot.OpponentPlayedCards[0].PlayCount != 2 {
		t.Fatalf("unexpected aggregated play count: %#v", snapshot.OpponentPlayedCards[0])
	}

	mustOverlayProject(t, projector, game.NewGameStartEvent("game-2", state.GameModeStandard))
	if len(projector.Snapshot().OpponentPlayedCards) != 0 {
		t.Fatalf("expected opponent cards reset for new match")
	}
}

func TestOverlaySnapshotKeepsOpponentCardsInFirstPlayChronology(t *testing.T) {
	t.Parallel()

	projector := NewOverlaySnapshotProjector(OverlaySnapshotProjectorOptions{
		Now: fixedTimeSequence(
			time.Unix(100, 0),
			time.Unix(101, 0),
			time.Unix(102, 0),
			time.Unix(103, 0),
		),
		CardDB: &stubLiveCardStatsCardDB{
			cards: map[string]*carddb.CardInfo{
				"LATE":  {ID: "LATE", Name: "Поздняя карта", Cost: 1},
				"EARLY": {ID: "EARLY", Name: "Ранняя карта", Cost: 9},
			},
		},
	})

	mustOverlayProject(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustOverlayProject(t, projector, game.NewCardPlayedEvent(types.PlayerTypeOpponent, card.New("LATE", 1, types.PlayerTypeOpponent), 0))
	mustOverlayProject(t, projector, game.NewCardPlayedEvent(types.PlayerTypeOpponent, card.New("EARLY", 2, types.PlayerTypeOpponent), 0))
	mustOverlayProject(t, projector, game.NewCardPlayedEvent(types.PlayerTypeOpponent, card.New("LATE", 3, types.PlayerTypeOpponent), 0))

	snapshot := projector.Snapshot()
	if len(snapshot.OpponentPlayedCards) != 2 {
		t.Fatalf("unexpected opponent card count: %#v", snapshot.OpponentPlayedCards)
	}
	if snapshot.OpponentPlayedCards[0].CardID != "LATE" {
		t.Fatalf("expected first-play chronology to keep LATE first, got %#v", snapshot.OpponentPlayedCards)
	}
	if snapshot.OpponentPlayedCards[1].CardID != "EARLY" {
		t.Fatalf("expected EARLY second, got %#v", snapshot.OpponentPlayedCards)
	}
}

func TestOverlaySnapshotRestorePreservesCounts(t *testing.T) {
	t.Parallel()

	original := NewOverlaySnapshotProjector(OverlaySnapshotProjectorOptions{})
	mustOverlayProject(t, original, game.NewDeckListEvent(types.PlayerTypePlayer, []*card.Card{
		card.New("DECK_1", 1, types.PlayerTypePlayer),
		card.New("DECK_1", 2, types.PlayerTypePlayer),
	}, "Deck", "deck-1", ""))
	mustOverlayProject(t, original, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustOverlayProject(t, original, game.NewCardDrawnEvent(types.PlayerTypePlayer, card.New("DECK_1", 3, types.PlayerTypePlayer)))
	mustOverlayProject(t, original, game.NewCardPlayedEvent(types.PlayerTypePlayer, card.New("DECK_1", 4, types.PlayerTypePlayer), 0))

	persisted := original.PersistenceSnapshot()
	if persisted == nil {
		t.Fatalf("expected persistence snapshot")
	}

	sink := &recordingOverlaySink{}
	resumed := NewOverlaySnapshotProjector(OverlaySnapshotProjectorOptions{Sink: sink})
	if err := resumed.Restore(*persisted); err != nil {
		t.Fatalf("restore snapshot: %v", err)
	}

	snapshot := resumed.Snapshot()
	if snapshot.PlayerDeckRemaining[0].RemainingCopies != 1 || snapshot.PlayerDeckRemaining[0].DrawnCopies != 1 || snapshot.PlayerDeckRemaining[0].PlayedCopies != 1 {
		t.Fatalf("unexpected restored counters: %#v", snapshot.PlayerDeckRemaining[0])
	}
	if len(sink.snapshots) != 1 {
		t.Fatalf("expected restore publish, got %d snapshots", len(sink.snapshots))
	}
}

func TestOverlaySnapshotHealthUpdatePublishesDegradedState(t *testing.T) {
	t.Parallel()

	sink := &recordingOverlaySink{}
	projector := NewOverlaySnapshotProjector(OverlaySnapshotProjectorOptions{Sink: sink})
	if err := projector.SetHealth(overlay.Health{
		Monitor:   overlay.HealthStateDegraded,
		Transport: overlay.HealthStateHealthy,
	}); err != nil {
		t.Fatalf("set health: %v", err)
	}

	snapshot := projector.Snapshot()
	if snapshot.Health.Monitor != overlay.HealthStateDegraded {
		t.Fatalf("unexpected monitor health: %#v", snapshot.Health)
	}
	if len(sink.snapshots) != 1 {
		t.Fatalf("expected health publish, got %d", len(sink.snapshots))
	}
}

func TestOverlaySnapshotFileSinkAtomicReplace(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "overlay_snapshot.json")
	sink, err := NewAtomicFileOverlaySnapshotSink(path)
	if err != nil {
		t.Fatalf("create sink: %v", err)
	}

	first := overlay.NewSnapshot()
	first.Match = overlay.Match{MatchID: "game:1", GameID: "1", Status: overlay.MatchStatusActive}
	second := first.Clone()
	second.Match.Status = overlay.MatchStatusFinalized

	if err := sink.Publish(first); err != nil {
		t.Fatalf("publish first snapshot: %v", err)
	}
	if err := sink.Publish(second); err != nil {
		t.Fatalf("publish second snapshot: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if strings.Contains(string(data), `"status": "active"`) {
		t.Fatalf("expected replaced file content, got:\n%s", string(data))
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("expected no temp file left behind, got %v", err)
	}
}

func TestOverlaySnapshotJSONContract(t *testing.T) {
	t.Parallel()

	projector := NewOverlaySnapshotProjector(OverlaySnapshotProjectorOptions{})
	mustOverlayProject(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))

	payload, err := marshalOverlaySnapshot(projector.Snapshot())
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}

	var decoded struct {
		SchemaVersion       string            `json:"schema_version"`
		PlayerDeckRemaining []json.RawMessage `json:"player_deck_remaining"`
		OpponentPlayedCards []json.RawMessage `json:"opponent_played_cards"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if decoded.SchemaVersion != overlay.SchemaVersionV1 {
		t.Fatalf("unexpected schema version: %q", decoded.SchemaVersion)
	}
	if len(decoded.PlayerDeckRemaining) != 0 || len(decoded.OpponentPlayedCards) != 0 {
		t.Fatalf("unexpected initial arrays: %#v", decoded)
	}
}

func mustOverlayProject(t *testing.T, projector *OverlaySnapshotProjector, event game.Event) {
	t.Helper()
	if err := projector.Process(event); err != nil {
		t.Fatalf("process event %T: %v", event, err)
	}
}

func requireOverlayDeckCard(t *testing.T, snapshot overlay.Snapshot, cardID string) overlay.PlayerDeckCard {
	t.Helper()
	for _, current := range snapshot.PlayerDeckRemaining {
		if current.CardID == cardID {
			return current
		}
	}
	t.Fatalf("expected player deck card %q in snapshot: %#v", cardID, snapshot.PlayerDeckRemaining)
	return overlay.PlayerDeckCard{}
}
