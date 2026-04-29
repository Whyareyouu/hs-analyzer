package api

import (
	"fmt"
	"slices"
	"strings"

	"hsanalyzer/internal/carddb"
)

type DeckListItem struct {
	DeckID       string  `json:"deck_id"`
	Name         string  `json:"name"`
	Class        string  `json:"class"`
	Wins         int     `json:"wins"`
	Losses       int     `json:"losses"`
	Winrate      float64 `json:"winrate"`
	LastPlayedAt string  `json:"last_played_at,omitempty"`
	Icon         string  `json:"icon,omitempty"`
}

type MatchupEntry struct {
	OpponentClass     string  `json:"opponent_class,omitempty"`
	OpponentArchetype string  `json:"opponent_archetype,omitempty"`
	Wins              int     `json:"wins"`
	Losses            int     `json:"losses"`
	Winrate           float64 `json:"winrate"`
	Matches           int     `json:"matches"`
}

type DeckOverview struct {
	DeckID        string         `json:"deck_id"`
	Name          string         `json:"name"`
	Class         string         `json:"class"`
	Wins          int            `json:"wins"`
	Losses        int            `json:"losses"`
	TotalMatches  int            `json:"total_matches"`
	Winrate       float64        `json:"winrate"`
	BestMatchups  []MatchupEntry `json:"best_matchups"`
	WorstMatchups []MatchupEntry `json:"worst_matchups"`
}

type DeckCard struct {
	CardID      string `json:"card_id"`
	Name        string `json:"name"`
	ManaCost    int    `json:"mana_cost"`
	Copies      int    `json:"copies"`
	Rarity      string `json:"rarity,omitempty"`
	IsLegendary bool   `json:"is_legendary"`
	Image       string `json:"image,omitempty"`
}

type DashboardStore interface {
	ListDecks() []DeckListItem
	GetDeckOverview(deckID string) (DeckOverview, bool)
	GetDeckCards(deckID string) ([]DeckCard, bool)
	GetDeckMatchups(deckID string) ([]MatchupEntry, bool)
}

type MemoryDashboardStore struct {
	decks     []DeckListItem
	overviews map[string]DeckOverview
	cards     map[string][]DeckCard
	matchups  map[string][]MatchupEntry
}

func NewMemoryDashboardStore() *MemoryDashboardStore {
	deckID := "deck-spell-dh"
	decks := []DeckListItem{
		{
			DeckID:       deckID,
			Name:         "Spell Demon Hunter",
			Class:        string(carddb.CardClassDemonHunter),
			Wins:         18,
			Losses:       12,
			Winrate:      0.60,
			LastPlayedAt: "2026-04-27T16:30:00Z",
			Icon:         "DH",
		},
		{
			DeckID:       "deck-rainbow-dk",
			Name:         "Rainbow Death Knight",
			Class:        string(carddb.CardClassDeathKnight),
			Wins:         11,
			Losses:       9,
			Winrate:      0.55,
			LastPlayedAt: "2026-04-25T19:12:00Z",
			Icon:         "DK",
		},
	}

	store := &MemoryDashboardStore{
		decks:     decks,
		overviews: make(map[string]DeckOverview, len(decks)),
		cards:     make(map[string][]DeckCard, len(decks)),
		matchups:  make(map[string][]MatchupEntry, len(decks)),
	}

	store.overviews[deckID] = DeckOverview{
		DeckID:       deckID,
		Name:         "Spell Demon Hunter",
		Class:        string(carddb.CardClassDemonHunter),
		Wins:         18,
		Losses:       12,
		TotalMatches: 30,
		Winrate:      0.60,
		BestMatchups: []MatchupEntry{
			{OpponentClass: string(carddb.CardClassWarrior), Wins: 6, Losses: 2, Matches: 8, Winrate: 0.75},
			{OpponentClass: string(carddb.CardClassDruid), Wins: 5, Losses: 2, Matches: 7, Winrate: 0.71},
			{OpponentClass: string(carddb.CardClassHunter), Wins: 4, Losses: 2, Matches: 6, Winrate: 0.67},
		},
		WorstMatchups: []MatchupEntry{
			{OpponentClass: string(carddb.CardClassMage), Wins: 1, Losses: 4, Matches: 5, Winrate: 0.20},
			{OpponentClass: string(carddb.CardClassPriest), Wins: 2, Losses: 4, Matches: 6, Winrate: 0.33},
			{OpponentClass: string(carddb.CardClassShaman), Wins: 0, Losses: 2, Matches: 2, Winrate: 0},
		},
	}
	store.cards[deckID] = []DeckCard{
		newDeckCard("CARD_001", "Taste of Chaos", 1, 2, carddb.RarityCommon, false),
		newDeckCard("CARD_002", "Parched Desperado", 2, 2, carddb.RarityRare, false),
		newDeckCard("CARD_003", "Illidari Studies", 1, 2, carddb.RarityCommon, false),
		newDeckCard("CARD_004", "Argus, the Emerald Star", 5, 1, carddb.RarityLegendary, true),
	}
	store.matchups[deckID] = append([]MatchupEntry(nil), store.overviews[deckID].BestMatchups...)
	store.matchups[deckID] = append(store.matchups[deckID], store.overviews[deckID].WorstMatchups...)

	secondaryID := "deck-rainbow-dk"
	store.overviews[secondaryID] = DeckOverview{
		DeckID:       secondaryID,
		Name:         "Rainbow Death Knight",
		Class:        string(carddb.CardClassDeathKnight),
		Wins:         11,
		Losses:       9,
		TotalMatches: 20,
		Winrate:      0.55,
		BestMatchups: []MatchupEntry{
			{OpponentClass: string(carddb.CardClassPaladin), Wins: 4, Losses: 1, Matches: 5, Winrate: 0.80},
			{OpponentClass: string(carddb.CardClassRogue), Wins: 3, Losses: 1, Matches: 4, Winrate: 0.75},
		},
		WorstMatchups: []MatchupEntry{
			{OpponentClass: string(carddb.CardClassDruid), Wins: 1, Losses: 3, Matches: 4, Winrate: 0.25},
		},
	}
	store.cards[secondaryID] = []DeckCard{
		newDeckCard("CARD_101", "Skeletal Sidekick", 1, 2, carddb.RarityCommon, false),
		newDeckCard("CARD_102", "Rainbow Seamstress", 3, 2, carddb.RarityEpic, false),
		newDeckCard("CARD_103", "The Primus", 8, 1, carddb.RarityLegendary, true),
	}
	store.matchups[secondaryID] = append([]MatchupEntry(nil), store.overviews[secondaryID].BestMatchups...)
	store.matchups[secondaryID] = append(store.matchups[secondaryID], store.overviews[secondaryID].WorstMatchups...)

	return store
}

func (s *MemoryDashboardStore) ListDecks() []DeckListItem {
	if s == nil {
		return nil
	}
	result := append([]DeckListItem(nil), s.decks...)
	slices.SortFunc(result, func(left, right DeckListItem) int {
		switch {
		case left.LastPlayedAt > right.LastPlayedAt:
			return -1
		case left.LastPlayedAt < right.LastPlayedAt:
			return 1
		case left.Name < right.Name:
			return -1
		case left.Name > right.Name:
			return 1
		default:
			return 0
		}
	})
	return result
}

func (s *MemoryDashboardStore) GetDeckOverview(deckID string) (DeckOverview, bool) {
	if s == nil {
		return DeckOverview{}, false
	}
	overview, ok := s.overviews[strings.TrimSpace(deckID)]
	if !ok {
		return DeckOverview{}, false
	}
	overview.BestMatchups = append([]MatchupEntry(nil), overview.BestMatchups...)
	overview.WorstMatchups = append([]MatchupEntry(nil), overview.WorstMatchups...)
	return overview, true
}

func (s *MemoryDashboardStore) GetDeckCards(deckID string) ([]DeckCard, bool) {
	if s == nil {
		return nil, false
	}
	cards, ok := s.cards[strings.TrimSpace(deckID)]
	if !ok {
		return nil, false
	}
	result := append([]DeckCard(nil), cards...)
	slices.SortFunc(result, func(left, right DeckCard) int {
		switch {
		case left.ManaCost < right.ManaCost:
			return -1
		case left.ManaCost > right.ManaCost:
			return 1
		case left.Name < right.Name:
			return -1
		case left.Name > right.Name:
			return 1
		default:
			return 0
		}
	})
	return result, true
}

func (s *MemoryDashboardStore) GetDeckMatchups(deckID string) ([]MatchupEntry, bool) {
	if s == nil {
		return nil, false
	}
	matchups, ok := s.matchups[strings.TrimSpace(deckID)]
	if !ok {
		return nil, false
	}
	result := append([]MatchupEntry(nil), matchups...)
	slices.SortFunc(result, func(left, right MatchupEntry) int {
		switch {
		case left.Winrate > right.Winrate:
			return -1
		case left.Winrate < right.Winrate:
			return 1
		case matchupLabel(left) < matchupLabel(right):
			return -1
		case matchupLabel(left) > matchupLabel(right):
			return 1
		default:
			return 0
		}
	})
	return result, true
}

func newDeckCard(cardID, name string, manaCost, copies int, rarity carddb.Rarity, legendary bool) DeckCard {
	return DeckCard{
		CardID:      cardID,
		Name:        name,
		ManaCost:    manaCost,
		Copies:      copies,
		Rarity:      string(rarity),
		IsLegendary: legendary,
		Image:       fmt.Sprintf("https://art.hearthstonejson.com/v1/render/latest/ruRU/256x/%s.png", cardID),
	}
}

func matchupLabel(entry MatchupEntry) string {
	if strings.TrimSpace(entry.OpponentArchetype) != "" {
		return entry.OpponentArchetype
	}
	return entry.OpponentClass
}
