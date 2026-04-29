package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/parser"
)

type DashboardDeckCatalogDeck struct {
	DeckID string
	Name   string
	Class  string
	Icon   string
	Cards  []DeckCard
}

type DashboardDeckCatalogSource interface {
	LoadDeckCatalog() ([]DashboardDeckCatalogDeck, error)
}

type FileDeckCatalogSource struct {
	deckLogPath string
	parser      *parser.DeckParser
}

func NewFileDeckCatalogSource(logsPath, cardsDBPath string) (*FileDeckCatalogSource, error) {
	logsPath = strings.TrimSpace(logsPath)
	cardsDBPath = strings.TrimSpace(cardsDBPath)
	if logsPath == "" {
		return nil, fmt.Errorf("logs path is empty")
	}
	if cardsDBPath == "" {
		return nil, fmt.Errorf("cards database path is empty")
	}

	deckLogPath := logsPath
	if info, err := os.Stat(logsPath); err == nil && info.IsDir() {
		deckLogPath = filepath.Join(logsPath, "Decks.log")
	}

	db := carddb.NewJSONDatabase()
	if err := db.Load(cardsDBPath); err != nil {
		return nil, fmt.Errorf("load cards database: %w", err)
	}

	return &FileDeckCatalogSource{
		deckLogPath: filepath.Clean(deckLogPath),
		parser:      parser.NewDeckParser(db),
	}, nil
}

func (s *FileDeckCatalogSource) LoadDeckCatalog() ([]DashboardDeckCatalogDeck, error) {
	if s == nil || s.parser == nil {
		return []DashboardDeckCatalogDeck{}, nil
	}

	file, err := os.Open(s.deckLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []DashboardDeckCatalogDeck{}, nil
		}
		return nil, fmt.Errorf("open deck catalog log: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	entries, _, err := s.parser.ParseCatalog(context.Background(), file)
	if err != nil {
		return nil, fmt.Errorf("parse deck catalog: %w", err)
	}

	uniqueByKey := make(map[string]DashboardDeckCatalogDeck, len(entries))
	order := make([]string, 0, len(entries))
	for _, entry := range entries {
		deckID := strings.TrimSpace(entry.DeckID)
		name := strings.TrimSpace(entry.Name)
		if deckID == "" {
			deckID = name
		}
		if deckID == "" {
			continue
		}
		if name == "" {
			name = deckID
		}

		className := fallbackDeckClass(deriveDeckClassFromCards(entry.Cards))
		deck := DashboardDeckCatalogDeck{
			DeckID: deckID,
			Name:   name,
			Class:  className,
			Icon:   deckIcon(className),
			Cards:  buildDeckCardsFromCards(entry.Cards),
		}

		uniqueKey := strings.TrimSpace(entry.Code)
		if uniqueKey == "" {
			uniqueKey = deckID
		}
		if _, exists := uniqueByKey[uniqueKey]; !exists {
			order = append(order, uniqueKey)
		}
		uniqueByKey[uniqueKey] = deck
	}

	result := make([]DashboardDeckCatalogDeck, 0, len(order))
	for _, key := range order {
		result = append(result, uniqueByKey[key])
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func buildDeckCardsFromCards(cards []*card.Card) []DeckCard {
	type aggregate struct {
		card DeckCard
	}

	grouped := make(map[string]*aggregate)
	order := make([]string, 0, len(cards))
	for _, current := range cards {
		if current == nil {
			continue
		}
		cardID := strings.TrimSpace(current.ID)
		if cardID == "" {
			continue
		}
		entry, ok := grouped[cardID]
		if !ok {
			entry = &aggregate{card: deckCardFromCard(current)}
			grouped[cardID] = entry
			order = append(order, cardID)
		}
		entry.card.Copies++
	}

	result := make([]DeckCard, 0, len(order))
	for _, cardID := range order {
		result = append(result, grouped[cardID].card)
	}
	return result
}

func deckCardFromCard(current *card.Card) DeckCard {
	result := DeckCard{
		CardID: strings.TrimSpace(current.ID),
	}
	if current != nil {
		result.ManaCost = current.Cost
	}

	info := cardInfoFromDeck(current)
	if info != nil {
		result.Name = strings.TrimSpace(info.Name)
		result.ManaCost = info.Cost
		result.Rarity = string(info.Rarity)
		result.IsLegendary = info.Rarity == carddb.RarityLegendary
		if result.CardID != "" {
			result.Image = fmt.Sprintf("https://art.hearthstonejson.com/v1/render/latest/ruRU/256x/%s.png", result.CardID)
		}
	}
	if result.Name == "" {
		result.Name = result.CardID
	}
	return result
}

func deriveDeckClassFromCards(cards []*card.Card) string {
	for _, current := range cards {
		info := cardInfoFromDeck(current)
		if info == nil {
			continue
		}
		if info.Class != "" && info.Class != carddb.CardClassNeutral {
			return string(info.Class)
		}
	}
	return ""
}

func cardInfoFromDeck(current *card.Card) *carddb.CardInfo {
	if current == nil || current.Info == nil {
		return nil
	}
	switch info := current.Info.(type) {
	case *carddb.CardInfo:
		return info
	case carddb.CardInfo:
		return &info
	default:
		return nil
	}
}
