package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileDeckCatalogSourceLoadsRealDecksLog(t *testing.T) {
	t.Parallel()

	source, err := NewFileDeckCatalogSource(
		filepath.Join("..", "..", "..", "..", "logs"),
		filepath.Join("..", "..", "..", "..", "data", "cards.json"),
	)
	if err != nil {
		t.Fatalf("create deck catalog source: %v", err)
	}

	decks, err := source.LoadDeckCatalog()
	if err != nil {
		t.Fatalf("load deck catalog: %v", err)
	}
	if len(decks) == 0 {
		t.Fatalf("expected decks from real Decks.log")
	}

	var selectedDeck *DashboardDeckCatalogDeck
	for index := range decks {
		if decks[index].DeckID == "2732587110" {
			selectedDeck = &decks[index]
			break
		}
	}
	if selectedDeck == nil {
		t.Fatalf("expected selected deck 2732587110 in catalog")
	}
	if selectedDeck.Name == "" {
		t.Fatalf("expected deck name in catalog entry")
	}
	totalCopies := 0
	for _, current := range selectedDeck.Cards {
		totalCopies += current.Copies
	}
	if totalCopies != 30 {
		t.Fatalf("expected 30 total copies in selected catalog deck, got %d", totalCopies)
	}
	if selectedDeck.Class == "UNKNOWN" {
		t.Fatalf("expected resolved deck class")
	}
}

func TestFileDeckCatalogSourceDeduplicatesDecksByCode(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	const deckCode = "AAECAZ6rBAaEnQe3rQePsQfQvwfOyQeb1AcM4+YGqvwGq/wG94MH6IcH0pcH0LIH7LIHhL0HtcAHr8EHoMUHAAA="
	deckLog := []byte(
		"I 00:00:00.0000000 ### First Name\n" +
			"I 00:00:00.0000000 # Deck ID: 1\n" +
			"I 00:00:00.0000000 " + deckCode + "\n" +
			"I 00:00:01.0000000 ### Latest Name\n" +
			"I 00:00:01.0000000 # Deck ID: 2\n" +
			"I 00:00:01.0000000 " + deckCode + "\n",
	)
	if err := os.WriteFile(filepath.Join(tempDir, "Decks.log"), deckLog, 0o644); err != nil {
		t.Fatalf("write Decks.log: %v", err)
	}

	source, err := NewFileDeckCatalogSource(tempDir, filepath.Join("..", "..", "..", "..", "data", "cards.json"))
	if err != nil {
		t.Fatalf("create source: %v", err)
	}

	decks, err := source.LoadDeckCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	if len(decks) != 1 {
		t.Fatalf("expected one unique deck, got %d", len(decks))
	}
	if decks[0].DeckID != "2" || decks[0].Name != "Latest Name" {
		t.Fatalf("expected latest duplicate deck metadata, got %#v", decks[0])
	}
}

func TestMatchDashboardStoreUsesCatalogWhenNoMatches(t *testing.T) {
	t.Parallel()

	store := NewMatchDashboardStore(NewMemoryStore(), fakeDeckCatalogSource{
		decks: []DashboardDeckCatalogDeck{
			{
				DeckID: "deck-1",
				Name:   "Spell Demon Hunter",
				Class:  "DEMONHUNTER",
				Icon:   "DH",
				Cards: []DeckCard{
					{CardID: "CARD_1", Name: "Taste of Chaos", ManaCost: 1, Copies: 2},
				},
			},
		},
	})

	decks := store.ListDecks()
	if len(decks) != 1 {
		t.Fatalf("expected one catalog deck, got %d", len(decks))
	}
	if decks[0].DeckID != "deck-1" {
		t.Fatalf("unexpected deck id: %s", decks[0].DeckID)
	}
	if decks[0].Wins != 0 || decks[0].Losses != 0 {
		t.Fatalf("expected zero stats for catalog-only deck: %#v", decks[0])
	}

	overview, ok := store.GetDeckOverview("deck-1")
	if !ok {
		t.Fatalf("expected deck overview")
	}
	if overview.TotalMatches != 0 {
		t.Fatalf("expected zero total matches, got %d", overview.TotalMatches)
	}

	cards, ok := store.GetDeckCards("deck-1")
	if !ok || len(cards) != 1 {
		t.Fatalf("expected catalog-backed cards, got ok=%t cards=%d", ok, len(cards))
	}

	matchups, ok := store.GetDeckMatchups("deck-1")
	if !ok {
		t.Fatalf("expected empty matchups slice")
	}
	if len(matchups) != 0 {
		t.Fatalf("expected empty matchups, got %d", len(matchups))
	}
}

func TestMatchDashboardStoreMergesCatalogAndMatchStats(t *testing.T) {
	t.Parallel()

	matchStore := NewMemoryStore()
	payload := backendDashboardPayload()
	if _, err := matchStore.SaveMatch(payload, "match-1"); err != nil {
		t.Fatalf("save match: %v", err)
	}

	store := NewMatchDashboardStore(matchStore, fakeDeckCatalogSource{
		decks: []DashboardDeckCatalogDeck{
			{
				DeckID: "deck-spell-dh",
				Name:   "Spell Demon Hunter",
				Class:  "DEMONHUNTER",
				Icon:   "DH",
				Cards: []DeckCard{
					{CardID: "CARD_0", Name: "Catalog Card", ManaCost: 0, Copies: 2},
				},
			},
		},
	})

	decks := store.ListDecks()
	if len(decks) != 1 {
		t.Fatalf("expected one merged deck, got %d", len(decks))
	}
	if decks[0].Wins != 1 || decks[0].Losses != 0 {
		t.Fatalf("expected uploaded stats to win merge: %#v", decks[0])
	}

	cards, ok := store.GetDeckCards("deck-spell-dh")
	if !ok {
		t.Fatalf("expected merged deck cards")
	}
	if len(cards) != 1 || cards[0].CardID != "CARD_1" {
		t.Fatalf("expected match payload cards to override catalog cards: %#v", cards)
	}

	overview, ok := store.GetDeckOverview("deck-spell-dh")
	if !ok {
		t.Fatalf("expected merged overview")
	}
	if overview.Winrate != 1 {
		t.Fatalf("expected winrate 1.0, got %f", overview.Winrate)
	}
}

type fakeDeckCatalogSource struct {
	decks []DashboardDeckCatalogDeck
	err   error
}

func (s fakeDeckCatalogSource) LoadDeckCatalog() ([]DashboardDeckCatalogDeck, error) {
	if s.err != nil {
		return nil, s.err
	}
	return append([]DashboardDeckCatalogDeck(nil), s.decks...), nil
}

var _ DashboardDeckCatalogSource = fakeDeckCatalogSource{}
