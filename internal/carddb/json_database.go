package carddb

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// JSONDatabase реализует CardDatabase на основе JSON-файла.
type JSONDatabase struct {
	mu           sync.RWMutex
	cardsByID    map[string]*CardInfo
	cardsByDBFID map[int]*CardInfo
	cardsByName  map[string]*CardInfo
}

// NewJSONDatabase создает новую JSON-базу карт.
func NewJSONDatabase() *JSONDatabase {
	return &JSONDatabase{
		cardsByID:    make(map[string]*CardInfo),
		cardsByDBFID: make(map[int]*CardInfo),
		cardsByName:  make(map[string]*CardInfo),
	}
}

// Load загружает базу карт из JSON.
func (db *JSONDatabase) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read cards file: %w", err)
	}

	loadedCards, err := parseCards(data)
	if err != nil {
		return err
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	db.cardsByID = make(map[string]*CardInfo, len(loadedCards))
	db.cardsByDBFID = make(map[int]*CardInfo, len(loadedCards))
	db.cardsByName = make(map[string]*CardInfo, len(loadedCards)*2)

	for i := range loadedCards {
		card := loadedCards[i]
		if card.ID == "" {
			continue
		}

		cardCopy := card
		db.cardsByID[card.ID] = &cardCopy
		if cardCopy.DBFID > 0 {
			db.cardsByDBFID[cardCopy.DBFID] = &cardCopy
		}
		db.indexCardName(cardCopy.Name, &cardCopy)
	}

	return nil
}

// GetCard возвращает карту по ID.
func (db *JSONDatabase) GetCard(id string) (*CardInfo, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	card, ok := db.cardsByID[id]
	return card, ok
}

// GetCardByDBFID возвращает карту по DBF ID.
func (db *JSONDatabase) GetCardByDBFID(dbfID int) (*CardInfo, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	card, ok := db.cardsByDBFID[dbfID]
	return card, ok
}

// GetCardByName возвращает карту по имени (регистронезависимо).
func (db *JSONDatabase) GetCardByName(name string) (*CardInfo, bool) {
	key := normalizeName(name)
	if key == "" {
		return nil, false
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	card, ok := db.cardsByName[key]
	return card, ok
}

func (db *JSONDatabase) indexCardName(name string, card *CardInfo) {
	key := normalizeName(name)
	if key != "" {
		db.cardsByName[key] = card
	}
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func parseCards(data []byte) ([]CardInfo, error) {
	var cards []CardInfo
	if err := json.Unmarshal(data, &cards); err == nil {
		return cards, nil
	}

	var cardsMap map[string]CardInfo
	if err := json.Unmarshal(data, &cardsMap); err == nil {
		result := make([]CardInfo, 0, len(cardsMap))
		for id, c := range cardsMap {
			if c.ID == "" {
				c.ID = id
			}
			result = append(result, c)
		}
		return result, nil
	}

	return nil, fmt.Errorf("parse cards json: unsupported format")
}
