package api

import (
	"sort"
	"strings"

	"hsanalyzer/pkg/contracts/matchupload"
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

type DeckHistoryEntry struct {
	MatchID       string `json:"match_id"`
	OpponentClass string `json:"opponent_class,omitempty"`
	Result        string `json:"result,omitempty"`
	PlayedAt      string `json:"played_at,omitempty"`
}

type DashboardStore interface {
	ListDecks() []DeckListItem
	GetDeckOverview(deckID string) (DeckOverview, bool)
	GetDeckCards(deckID string) ([]DeckCard, bool)
	GetDeckMatchups(deckID string) ([]MatchupEntry, bool)
	GetDeckHistory(deckID string) ([]DeckHistoryEntry, bool)
}

type MatchDashboardStore struct {
	store         MatchStore
	catalogSource DashboardDeckCatalogSource
}

func NewMatchDashboardStore(store MatchStore, catalogSources ...DashboardDeckCatalogSource) *MatchDashboardStore {
	dashboardStore := &MatchDashboardStore{store: store}
	if len(catalogSources) > 0 {
		dashboardStore.catalogSource = catalogSources[0]
	}
	return dashboardStore
}

type dashboardDeckAggregate struct {
	listItem   DeckListItem
	overview   DeckOverview
	cards      []DeckCard
	matchups   map[string]*MatchupEntry
	history    []DeckHistoryEntry
	lastPlayed string
}

type dashboardSnapshot struct {
	decks map[string]*dashboardDeckAggregate
}

func (s *MatchDashboardStore) ListDecks() []DeckListItem {
	snapshot := s.buildSnapshot()
	result := make([]DeckListItem, 0, len(snapshot.decks))
	for _, current := range snapshot.decks {
		result = append(result, current.listItem)
	}
	sort.Slice(result, func(i, j int) bool {
		switch {
		case result[i].LastPlayedAt > result[j].LastPlayedAt:
			return true
		case result[i].LastPlayedAt < result[j].LastPlayedAt:
			return false
		default:
			return result[i].Name < result[j].Name
		}
	})
	return result
}

func (s *MatchDashboardStore) GetDeckOverview(deckID string) (DeckOverview, bool) {
	current, ok := s.buildSnapshot().decks[strings.TrimSpace(deckID)]
	if !ok {
		return DeckOverview{}, false
	}
	overview := current.overview
	best, worst := buildMatchupSlices(current.matchups)
	overview.BestMatchups = best
	overview.WorstMatchups = worst
	return overview, true
}

func (s *MatchDashboardStore) GetDeckCards(deckID string) ([]DeckCard, bool) {
	current, ok := s.buildSnapshot().decks[strings.TrimSpace(deckID)]
	if !ok {
		return nil, false
	}
	result := append([]DeckCard(nil), current.cards...)
	sort.Slice(result, func(i, j int) bool {
		switch {
		case result[i].ManaCost < result[j].ManaCost:
			return true
		case result[i].ManaCost > result[j].ManaCost:
			return false
		default:
			return result[i].Name < result[j].Name
		}
	})
	return result, true
}

func (s *MatchDashboardStore) GetDeckMatchups(deckID string) ([]MatchupEntry, bool) {
	current, ok := s.buildSnapshot().decks[strings.TrimSpace(deckID)]
	if !ok {
		return nil, false
	}
	result, _ := buildMatchupSlices(current.matchups)
	return result, true
}

func (s *MatchDashboardStore) GetDeckHistory(deckID string) ([]DeckHistoryEntry, bool) {
	current, ok := s.buildSnapshot().decks[strings.TrimSpace(deckID)]
	if !ok {
		return nil, false
	}
	result := append([]DeckHistoryEntry(nil), current.history...)
	sort.Slice(result, func(i, j int) bool {
		switch {
		case result[i].PlayedAt > result[j].PlayedAt:
			return true
		case result[i].PlayedAt < result[j].PlayedAt:
			return false
		default:
			return result[i].MatchID > result[j].MatchID
		}
	})
	return result, true
}

func (s *MatchDashboardStore) buildSnapshot() dashboardSnapshot {
	snapshot := dashboardSnapshot{decks: map[string]*dashboardDeckAggregate{}}
	if s == nil {
		return snapshot
	}

	if s.catalogSource != nil {
		if catalogDecks, err := s.catalogSource.LoadDeckCatalog(); err == nil {
			for _, catalogDeck := range catalogDecks {
				deckID := strings.TrimSpace(catalogDeck.DeckID)
				if deckID == "" {
					continue
				}
				snapshot.decks[deckID] = &dashboardDeckAggregate{
					listItem: DeckListItem{
						DeckID: deckID,
						Name:   strings.TrimSpace(catalogDeck.Name),
						Class:  fallbackDeckClass(catalogDeck.Class),
						Icon:   strings.TrimSpace(catalogDeck.Icon),
					},
					overview: DeckOverview{
						DeckID: deckID,
						Name:   strings.TrimSpace(catalogDeck.Name),
						Class:  fallbackDeckClass(catalogDeck.Class),
					},
					cards:    append([]DeckCard(nil), catalogDeck.Cards...),
					matchups: make(map[string]*MatchupEntry),
				}
			}
		}
	}

	if s.store != nil {
		for _, stored := range s.store.ListMatches() {
			summary, ok, err := stored.Payload.GetDashboardSummary()
			if err != nil || !ok {
				continue
			}

			deckID := dashboardDeckKey(summary)
			if deckID == "" {
				continue
			}

			current, exists := snapshot.decks[deckID]
			if !exists {
				current = &dashboardDeckAggregate{
					listItem: DeckListItem{
						DeckID: deckID,
						Name:   fallbackDeckName(summary, deckID),
						Class:  fallbackDeckClass(summary.PlayerClass),
						Icon:   deckIcon(summary.PlayerClass),
					},
					overview: DeckOverview{
						DeckID: deckID,
						Name:   fallbackDeckName(summary, deckID),
						Class:  fallbackDeckClass(summary.PlayerClass),
					},
					matchups: make(map[string]*MatchupEntry),
				}
				snapshot.decks[deckID] = current
			}

			playedAt := payloadPlayedAt(stored.Payload)
			if playedAt >= current.lastPlayed {
				current.lastPlayed = playedAt
				current.listItem.LastPlayedAt = playedAt
				if cards := convertDashboardCards(summary.DeckCards); len(cards) > 0 {
					current.cards = cards
				}
				if strings.TrimSpace(summary.DeckName) != "" {
					current.listItem.Name = strings.TrimSpace(summary.DeckName)
					current.overview.Name = strings.TrimSpace(summary.DeckName)
				}
				if strings.TrimSpace(summary.PlayerClass) != "" {
					className := strings.TrimSpace(summary.PlayerClass)
					current.listItem.Class = className
					current.listItem.Icon = deckIcon(className)
					current.overview.Class = className
				}
			}

			current.overview.TotalMatches++
			switch strings.TrimSpace(summary.Result) {
			case matchupload.DashboardResultWin:
				current.listItem.Wins++
				current.overview.Wins++
			case matchupload.DashboardResultLoss:
				current.listItem.Losses++
				current.overview.Losses++
			}

			if opponentClass := strings.TrimSpace(summary.OpponentClass); opponentClass != "" {
				entry, ok := current.matchups[opponentClass]
				if !ok {
					entry = &MatchupEntry{OpponentClass: opponentClass}
					current.matchups[opponentClass] = entry
				}
				entry.Matches++
				switch strings.TrimSpace(summary.Result) {
				case matchupload.DashboardResultWin:
					entry.Wins++
				case matchupload.DashboardResultLoss:
					entry.Losses++
				}
				entry.Winrate = ratio(entry.Wins, entry.Matches)
			}

			current.history = append(current.history, DeckHistoryEntry{
				MatchID:       strings.TrimSpace(stored.Payload.MatchID),
				OpponentClass: fallbackDeckClass(summary.OpponentClass),
				Result:        strings.TrimSpace(summary.Result),
				PlayedAt:      playedAt,
			})
		}
	}

	for _, current := range snapshot.decks {
		if strings.TrimSpace(current.listItem.Name) == "" {
			current.listItem.Name = current.listItem.DeckID
		}
		if strings.TrimSpace(current.listItem.Class) == "" {
			current.listItem.Class = "UNKNOWN"
		}
		if strings.TrimSpace(current.listItem.Icon) == "" {
			current.listItem.Icon = deckIcon(current.listItem.Class)
		}
		if strings.TrimSpace(current.overview.Name) == "" {
			current.overview.Name = current.listItem.Name
		}
		if strings.TrimSpace(current.overview.Class) == "" {
			current.overview.Class = current.listItem.Class
		}
		current.listItem.Winrate = ratio(current.listItem.Wins, current.overview.TotalMatches)
		current.overview.Winrate = ratio(current.overview.Wins, current.overview.TotalMatches)
	}

	return snapshot
}

func buildMatchupSlices(matchups map[string]*MatchupEntry) ([]MatchupEntry, []MatchupEntry) {
	if len(matchups) == 0 {
		return []MatchupEntry{}, []MatchupEntry{}
	}

	items := make([]MatchupEntry, 0, len(matchups))
	for _, current := range matchups {
		items = append(items, *current)
	}
	sort.Slice(items, func(i, j int) bool {
		switch {
		case items[i].Winrate > items[j].Winrate:
			return true
		case items[i].Winrate < items[j].Winrate:
			return false
		default:
			return items[i].OpponentClass < items[j].OpponentClass
		}
	})

	best := append([]MatchupEntry(nil), items...)
	worst := append([]MatchupEntry(nil), items...)
	sort.Slice(worst, func(i, j int) bool {
		switch {
		case worst[i].Winrate < worst[j].Winrate:
			return true
		case worst[i].Winrate > worst[j].Winrate:
			return false
		default:
			return worst[i].OpponentClass < worst[j].OpponentClass
		}
	})

	return best, worst
}

func payloadPlayedAt(payload matchupload.Payload) string {
	if strings.TrimSpace(payload.EndedAt) != "" {
		return strings.TrimSpace(payload.EndedAt)
	}
	return strings.TrimSpace(payload.StartedAt)
}

func convertDashboardCards(cards []matchupload.DashboardDeckCard) []DeckCard {
	result := make([]DeckCard, 0, len(cards))
	for _, current := range cards {
		result = append(result, DeckCard{
			CardID:      strings.TrimSpace(current.CardID),
			Name:        strings.TrimSpace(current.Name),
			ManaCost:    current.ManaCost,
			Copies:      current.Copies,
			Rarity:      strings.TrimSpace(current.Rarity),
			IsLegendary: current.IsLegendary,
			Image:       strings.TrimSpace(current.Image),
		})
	}
	return result
}

func dashboardDeckKey(summary matchupload.DashboardSummary) string {
	if strings.TrimSpace(summary.DeckID) != "" {
		return strings.TrimSpace(summary.DeckID)
	}
	return strings.TrimSpace(summary.DeckName)
}

func fallbackDeckName(summary matchupload.DashboardSummary, deckID string) string {
	if strings.TrimSpace(summary.DeckName) != "" {
		return strings.TrimSpace(summary.DeckName)
	}
	if strings.TrimSpace(deckID) != "" {
		return strings.TrimSpace(deckID)
	}
	return "Unknown Deck"
}

func fallbackDeckClass(value string) string {
	if strings.TrimSpace(value) == "" {
		return "UNKNOWN"
	}
	return strings.TrimSpace(value)
}

func deckIcon(className string) string {
	switch strings.TrimSpace(className) {
	case "DEMONHUNTER":
		return "DH"
	case "DEATHKNIGHT":
		return "DK"
	case "PALADIN":
		return "PA"
	case "PRIEST":
		return "PR"
	case "WARLOCK":
		return "WL"
	case "WARRIOR":
		return "WR"
	case "HUNTER":
		return "HU"
	case "MAGE":
		return "MA"
	case "ROGUE":
		return "RO"
	case "SHAMAN":
		return "SH"
	case "DRUID":
		return "DR"
	default:
		return "HS"
	}
}

func ratio(numerator, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
