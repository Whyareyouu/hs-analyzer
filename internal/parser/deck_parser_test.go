package parser

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/types"
)

func TestDeckParserParseSelectedDeckFromRealLogs(t *testing.T) {
	t.Parallel()

	deckLogPath := filepath.Join("..", "..", "logs", "Decks.log")
	deckLogData, err := os.ReadFile(deckLogPath)
	if err != nil {
		t.Fatalf("read Decks.log: %v", err)
	}

	db := carddb.NewJSONDatabase()
	if err := db.Load(filepath.Join("..", "..", "data", "cards.json")); err != nil {
		t.Fatalf("load cards db: %v", err)
	}

	parser := NewDeckParser(db)
	events, err := parser.Parse(context.Background(), strings.NewReader(string(deckLogData)))
	if err != nil {
		t.Fatalf("parse deck log: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	deckEvent, ok := events[0].(*game.DeckListEvent)
	if !ok {
		t.Fatalf("unexpected event type: %T", events[0])
	}
	if deckEvent.Player != types.PlayerTypePlayer {
		t.Fatalf("unexpected player: %s", deckEvent.Player)
	}
	if deckEvent.Name != "ZA | LE | PA" {
		t.Fatalf("unexpected deck name: %s", deckEvent.Name)
	}
	if deckEvent.DeckID != "2732587110" {
		t.Fatalf("unexpected deck id: %s", deckEvent.DeckID)
	}
	if len(deckEvent.Cards) != 30 {
		t.Fatalf("unexpected deck size: %d", len(deckEvent.Cards))
	}
	if deckEvent.Cards[0] == nil || deckEvent.Cards[0].ID == "" {
		t.Fatalf("expected first card to be resolved")
	}
	if deckEvent.Cards[0].Info == nil {
		t.Fatalf("expected first card info to be attached")
	}
}

func TestDeckParserParseRequiresCardDatabase(t *testing.T) {
	t.Parallel()

	parser := NewDeckParser(nil)
	_, err := parser.Parse(context.Background(), strings.NewReader(findLogLine(t, "Decks.log", "Finding Game With Deck:")))
	if err == nil {
		t.Fatalf("expected card database error")
	}
}

func TestDeckParserParseCatalogFromRealLogs(t *testing.T) {
	t.Parallel()

	deckLogPath := filepath.Join("..", "..", "logs", "Decks.log")
	deckLogData, err := os.ReadFile(deckLogPath)
	if err != nil {
		t.Fatalf("read Decks.log: %v", err)
	}

	db := carddb.NewJSONDatabase()
	if err := db.Load(filepath.Join("..", "..", "data", "cards.json")); err != nil {
		t.Fatalf("load cards db: %v", err)
	}

	parser := NewDeckParser(db)
	catalog, selectedDeckID, err := parser.ParseCatalog(context.Background(), strings.NewReader(string(deckLogData)))
	if err != nil {
		t.Fatalf("parse deck catalog: %v", err)
	}
	if len(catalog) == 0 {
		t.Fatalf("expected at least one deck in catalog")
	}
	if selectedDeckID != "2732587110" {
		t.Fatalf("unexpected selected deck id: %s", selectedDeckID)
	}

	selectedCount := 0
	for _, current := range catalog {
		if current.Selected {
			selectedCount++
			if current.DeckID != selectedDeckID {
				t.Fatalf("selected deck mismatch: got %s want %s", current.DeckID, selectedDeckID)
			}
		}
		if len(current.Cards) != 30 {
			t.Fatalf("expected 30 cards for deck %s, got %d", current.DeckID, len(current.Cards))
		}
	}
	if selectedCount != 1 {
		t.Fatalf("expected exactly one selected deck, got %d", selectedCount)
	}
}

func TestDeckParserParseStreamEmitsSelectedDeckEvents(t *testing.T) {
	t.Parallel()

	deckLogPath := filepath.Join("..", "..", "logs", "Decks.log")
	deckLogData, err := os.ReadFile(deckLogPath)
	if err != nil {
		t.Fatalf("read Decks.log: %v", err)
	}

	db := carddb.NewJSONDatabase()
	if err := db.Load(filepath.Join("..", "..", "data", "cards.json")); err != nil {
		t.Fatalf("load cards db: %v", err)
	}

	parser := NewDeckParser(db)
	events, err := parser.ParseStream(context.Background(), strings.NewReader(string(deckLogData)))
	if err != nil {
		t.Fatalf("parse stream: %v", err)
	}
	if len(events) == 0 {
		t.Fatalf("expected incremental deck events")
	}

	lastEvent, ok := events[len(events)-1].(*game.DeckListEvent)
	if !ok {
		t.Fatalf("unexpected last event type: %T", events[len(events)-1])
	}
	if lastEvent.DeckID != "2732587110" {
		t.Fatalf("unexpected selected deck id: %s", lastEvent.DeckID)
	}
	if lastEvent.Cards[0] == nil || lastEvent.Cards[0].Info == nil {
		t.Fatalf("expected card metadata in incremental deck event")
	}
}
