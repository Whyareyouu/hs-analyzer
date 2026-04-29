package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/parser"
	"hsanalyzer/internal/types"
)

func TestLiveEventSideCorrectorInvertsSidesWhenSelectedDeckCardsAreMappedAsOpponent(t *testing.T) {
	t.Parallel()

	corrector := newLiveEventSideCorrector()
	corrector.Correct(game.NewDeckListEvent(types.PlayerTypePlayer, []*card.Card{
		card.New("MY_DECK_CARD", 1, types.PlayerTypePlayer),
	}, "My Deck", "deck-1", "code"))

	selfCard := card.New("MY_DECK_CARD", 10, types.PlayerTypeOpponent)
	selfPlay := game.NewCardPlayedEvent(types.PlayerTypeOpponent, selfCard, 0)
	correctedSelfPlay, ok := corrector.Correct(selfPlay).(*game.CardPlayedEvent)
	if !ok {
		t.Fatalf("unexpected corrected self event type: %T", correctedSelfPlay)
	}
	if correctedSelfPlay.Player != types.PlayerTypePlayer {
		t.Fatalf("expected selected deck card to establish local side, got %s", correctedSelfPlay.Player)
	}
	if correctedSelfPlay.Card.Owner != types.PlayerTypePlayer || correctedSelfPlay.Card.Controller != types.PlayerTypePlayer {
		t.Fatalf("expected card ownership to be corrected, got owner=%s controller=%s", correctedSelfPlay.Card.Owner, correctedSelfPlay.Card.Controller)
	}

	opponentCard := card.New("OPP_CARD", 11, types.PlayerTypePlayer)
	opponentPlay := game.NewCardPlayedEvent(types.PlayerTypePlayer, opponentCard, 0)
	correctedOpponentPlay, ok := corrector.Correct(opponentPlay).(*game.CardPlayedEvent)
	if !ok {
		t.Fatalf("unexpected corrected opponent event type: %T", correctedOpponentPlay)
	}
	if correctedOpponentPlay.Player != types.PlayerTypeOpponent {
		t.Fatalf("expected opposite parser side to become opponent, got %s", correctedOpponentPlay.Player)
	}
}

func TestLiveEventSideCorrectorKeepsSidesWhenSelectedDeckCardsAreMappedAsPlayer(t *testing.T) {
	t.Parallel()

	corrector := newLiveEventSideCorrector()
	corrector.Correct(game.NewDeckListEvent(types.PlayerTypePlayer, []*card.Card{
		card.New("MY_DECK_CARD", 1, types.PlayerTypePlayer),
	}, "My Deck", "deck-1", "code"))

	selfPlay := game.NewCardPlayedEvent(types.PlayerTypePlayer, card.New("MY_DECK_CARD", 10, types.PlayerTypePlayer), 0)
	correctedSelfPlay, ok := corrector.Correct(selfPlay).(*game.CardPlayedEvent)
	if !ok {
		t.Fatalf("unexpected corrected self event type: %T", correctedSelfPlay)
	}
	if correctedSelfPlay.Player != types.PlayerTypePlayer {
		t.Fatalf("expected selected deck card to remain player, got %s", correctedSelfPlay.Player)
	}

	opponentPlay := game.NewCardPlayedEvent(types.PlayerTypeOpponent, card.New("OPP_CARD", 11, types.PlayerTypeOpponent), 0)
	correctedOpponentPlay, ok := corrector.Correct(opponentPlay).(*game.CardPlayedEvent)
	if !ok {
		t.Fatalf("unexpected corrected opponent event type: %T", correctedOpponentPlay)
	}
	if correctedOpponentPlay.Player != types.PlayerTypeOpponent {
		t.Fatalf("expected opponent side to remain opponent, got %s", correctedOpponentPlay.Player)
	}
}

func TestLiveEventSideCorrectorDoesNotLearnSideFromDrawnDeckCards(t *testing.T) {
	t.Parallel()

	corrector := newLiveEventSideCorrector()
	corrector.Correct(game.NewDeckListEvent(types.PlayerTypePlayer, []*card.Card{
		card.New("MY_DECK_CARD", 1, types.PlayerTypePlayer),
	}, "My Deck", "deck-1", "code"))

	drawnCard := card.New("MY_DECK_CARD", 10, types.PlayerTypeOpponent)
	correctedDraw, ok := corrector.Correct(game.NewCardDrawnEvent(types.PlayerTypeOpponent, drawnCard)).(*game.CardDrawnEvent)
	if !ok {
		t.Fatalf("unexpected corrected draw event type: %T", correctedDraw)
	}
	if correctedDraw.Player != types.PlayerTypeOpponent {
		t.Fatalf("draw events must not establish side inversion, got %s", correctedDraw.Player)
	}

	opponentPlay := game.NewCardPlayedEvent(types.PlayerTypeOpponent, card.New("OPP_CARD", 11, types.PlayerTypeOpponent), 0)
	correctedOpponentPlay, ok := corrector.Correct(opponentPlay).(*game.CardPlayedEvent)
	if !ok {
		t.Fatalf("unexpected corrected opponent event type: %T", correctedOpponentPlay)
	}
	if correctedOpponentPlay.Player != types.PlayerTypeOpponent {
		t.Fatalf("expected opponent side to remain opponent after a drawn deck card, got %s", correctedOpponentPlay.Player)
	}
}

func TestLiveEventSideCorrectorUsesRealDeckLogToFixPowerLogSideInversion(t *testing.T) {
	t.Parallel()

	projectRoot := projectRootForTest(t)
	db := carddb.NewJSONDatabase()
	if err := db.Load(filepath.Join(projectRoot, "data", "cards.json")); err != nil {
		t.Fatalf("load card db: %v", err)
	}

	deckFile, err := os.Open(filepath.Join(projectRoot, "logs", "Decks.log"))
	if err != nil {
		t.Fatalf("open Decks.log: %v", err)
	}
	defer func() {
		_ = deckFile.Close()
	}()

	corrector := newLiveEventSideCorrector()
	deckEvents, err := parser.NewDeckParser(db).Parse(context.Background(), deckFile)
	if err != nil {
		t.Fatalf("parse Decks.log: %v", err)
	}
	for _, event := range deckEvents {
		corrector.Correct(event)
	}

	powerFile, err := os.Open(filepath.Join(projectRoot, "logs", "Power.log"))
	if err != nil {
		t.Fatalf("open Power.log: %v", err)
	}
	defer func() {
		_ = powerFile.Close()
	}()

	powerEvents, err := parser.NewPowerParser().Parse(context.Background(), powerFile)
	if err != nil {
		t.Fatalf("parse Power.log: %v", err)
	}

	knownDeckCardsSeen := 0
	for _, event := range powerEvents {
		played, ok := corrector.Correct(event).(*game.CardPlayedEvent)
		if !ok || played.Card == nil {
			continue
		}
		if played.Card.ID != "TLC_515" && played.Card.ID != "EDR_528" {
			continue
		}
		knownDeckCardsSeen++
		if played.Player != types.PlayerTypePlayer {
			t.Fatalf("expected real selected-deck card %s to be corrected to player, got %s", played.Card.ID, played.Player)
		}
	}
	if knownDeckCardsSeen == 0 {
		t.Fatalf("expected to see selected deck cards in real Power.log")
	}
}
