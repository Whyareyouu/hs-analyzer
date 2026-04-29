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
	if len(catalog) < 7 {
		t.Fatalf("expected all known decks from Decks.log, got %d", len(catalog))
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

func TestDeckParserParseCatalogFromNameAndCodePairs(t *testing.T) {
	t.Parallel()

	const deckCode = "AAECAZ6rBAaEnQe3rQePsQfQvwfOyQeb1AcM4+YGqvwGq/wG94MH6IcH0pcH0LIH7LIHhL0HtcAHr8EHoMUHAAA="
	input := strings.Join([]string{
		`I 00:00:00.0000000 "Название колоды"`,
		"I 00:00:00.0000000 " + deckCode,
	}, "\n")

	db := carddb.NewJSONDatabase()
	if err := db.Load(filepath.Join("..", "..", "data", "cards.json")); err != nil {
		t.Fatalf("load cards db: %v", err)
	}

	parser := NewDeckParser(db)
	catalog, selectedDeckID, err := parser.ParseCatalog(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse simple deck catalog: %v", err)
	}
	if len(catalog) != 1 {
		t.Fatalf("unexpected catalog size: %d", len(catalog))
	}
	if catalog[0].Name != "Название колоды" {
		t.Fatalf("unexpected deck name: %s", catalog[0].Name)
	}
	if catalog[0].DeckID != deckIDFromCode(deckCode) {
		t.Fatalf("unexpected generated deck id: %s", catalog[0].DeckID)
	}
	if selectedDeckID != catalog[0].DeckID {
		t.Fatalf("expected latest simple deck to be selected, got %s", selectedDeckID)
	}
	if len(catalog[0].Cards) != 30 {
		t.Fatalf("unexpected deck size: %d", len(catalog[0].Cards))
	}
}

func TestDeckParserParseCatalogSkipsDeckstringMetadata(t *testing.T) {
	t.Parallel()

	const deckCode = "AAECAZ6rBAaEnQe3rQePsQfQvwfOyQeb1AcM4+YGqvwGq/wG94MH6IcH0pcH0LIH7LIHhL0HtcAHr8EHoMUHAAA="
	input := strings.Join([]string{
		`I 00:00:00.0000000 ### Метовая колода`,
		`I 00:00:00.0000000 # Class: Paladin`,
		`I 00:00:00.0000000 # Format: Standard`,
		`I 00:00:00.0000000 # Deck ID: 42`,
		"I 00:00:00.0000000 " + deckCode,
	}, "\n")

	db := carddb.NewJSONDatabase()
	if err := db.Load(filepath.Join("..", "..", "data", "cards.json")); err != nil {
		t.Fatalf("load cards db: %v", err)
	}

	parser := NewDeckParser(db)
	catalog, selectedDeckID, err := parser.ParseCatalog(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse metadata deck catalog: %v", err)
	}
	if len(catalog) != 1 {
		t.Fatalf("unexpected catalog size: %d", len(catalog))
	}
	if catalog[0].Name != "Метовая колода" {
		t.Fatalf("unexpected deck name: %s", catalog[0].Name)
	}
	if catalog[0].DeckID != "42" || selectedDeckID != "42" {
		t.Fatalf("unexpected deck ids: deck=%s selected=%s", catalog[0].DeckID, selectedDeckID)
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
