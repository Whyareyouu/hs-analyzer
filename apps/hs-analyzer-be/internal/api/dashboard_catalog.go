package api

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/types"
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
	cardDB      carddb.CardDatabase
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
		cardDB:      db,
	}, nil
}

func (s *FileDeckCatalogSource) LoadDeckCatalog() ([]DashboardDeckCatalogDeck, error) {
	if s == nil || s.cardDB == nil {
		return []DashboardDeckCatalogDeck{}, nil
	}

	data, err := os.ReadFile(s.deckLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []DashboardDeckCatalogDeck{}, nil
		}
		return nil, fmt.Errorf("read deck catalog log: %w", err)
	}
	entries, err := parseDashboardDeckCatalog(string(data), s.cardDB)
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

type dashboardDeckCatalogEntry struct {
	Name   string
	DeckID string
	Code   string
	Cards  []*card.Card
}

type dashboardDeckLogEntry struct {
	Name   string
	DeckID string
	Code   string
}

func parseDashboardDeckCatalog(raw string, db carddb.CardDatabase) ([]dashboardDeckCatalogEntry, error) {
	if db == nil {
		return nil, fmt.Errorf("card database is nil")
	}

	entries, err := parseDashboardDeckLogCatalog(raw)
	if err != nil {
		return nil, err
	}

	result := make([]dashboardDeckCatalogEntry, 0, len(entries))
	for _, entry := range entries {
		cards, err := decodeDashboardDeckCards(entry.Code, db)
		if err != nil {
			return nil, fmt.Errorf("decode deck %s: %w", entry.DeckID, err)
		}
		result = append(result, dashboardDeckCatalogEntry{
			Name:   entry.Name,
			DeckID: entry.DeckID,
			Code:   entry.Code,
			Cards:  cards,
		})
	}

	return result, nil
}

func parseDashboardDeckLogCatalog(raw string) ([]dashboardDeckLogEntry, error) {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	entries := make(map[string]dashboardDeckLogEntry)
	order := make([]string, 0)
	currentTrigger := ""

	for i := 0; i < len(lines); i++ {
		line := dashboardDeckLogPayload(lines[i])
		if isDashboardDeckTriggerLine(line) {
			currentTrigger = line
			continue
		}

		if _, ok := parseDashboardDeckNameLine(line, currentTrigger != ""); !ok {
			continue
		}

		entry, consumed, err := parseDashboardDeckBlock(lines, i)
		if err != nil {
			return nil, err
		}
		if entry.DeckID != "" {
			if _, exists := entries[entry.DeckID]; !exists {
				order = append(order, entry.DeckID)
			}
			entries[entry.DeckID] = entry
		}
		if strings.Contains(currentTrigger, "Finding Game With Deck:") {
			currentTrigger = ""
		}
		i = consumed
	}

	result := make([]dashboardDeckLogEntry, 0, len(order))
	for _, deckID := range order {
		result = append(result, entries[deckID])
	}
	return result, nil
}

func parseDashboardDeckBlock(lines []string, start int) (dashboardDeckLogEntry, int, error) {
	entry := dashboardDeckLogEntry{}
	i := start
	for ; i < len(lines); i++ {
		line := dashboardDeckLogPayload(lines[i])
		if line == "" {
			break
		}
		if isDashboardDeckTriggerLine(line) {
			i--
			break
		}
		if entry.Name != "" {
			if _, ok := parseDashboardDeckNameLine(line, true); ok && entry.Code == "" {
				i--
				break
			}
		}
		if entry.Name == "" {
			name, ok := parseDashboardDeckNameLine(line, true)
			if ok {
				entry.Name = name
				continue
			}
		}
		if entry.DeckID == "" && strings.Contains(line, "# Deck ID: ") {
			entry.DeckID = strings.TrimSpace(strings.SplitN(line, "# Deck ID: ", 2)[1])
			continue
		}
		if entry.Code == "" && isDashboardDeckCodeLine(line) {
			fields := strings.Fields(line)
			entry.Code = fields[len(fields)-1]
			if entry.DeckID == "" {
				entry.DeckID = dashboardDeckIDFromCode(entry.Code)
			}
			break
		}
	}

	if entry.Name == "" && entry.DeckID != "" {
		entry.Name = entry.DeckID
	}
	if entry.DeckID == "" || entry.Code == "" {
		return dashboardDeckLogEntry{}, i, fmt.Errorf("incomplete deck block: %#v", entry)
	}

	return entry, i, nil
}

func parseDashboardDeckNameLine(line string, allowPlain bool) (string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || isDashboardDeckTriggerLine(line) || strings.Contains(line, "# Deck ID: ") || isDashboardDeckCodeLine(line) {
		return "", false
	}
	if strings.Contains(line, "### ") {
		return strings.TrimSpace(strings.SplitN(line, "### ", 2)[1]), true
	}
	if strings.HasPrefix(line, `"`) && strings.Count(line, `"`) >= 2 {
		rest := strings.TrimPrefix(line, `"`)
		name, _, _ := strings.Cut(rest, `"`)
		name = strings.TrimSpace(name)
		return name, name != ""
	}
	if firstQuote := strings.Index(line, `"`); firstQuote >= 0 {
		rest := line[firstQuote+1:]
		name, _, found := strings.Cut(rest, `"`)
		name = strings.TrimSpace(name)
		return name, found && name != ""
	}
	if allowPlain && !strings.HasPrefix(line, "#") {
		return strings.TrimSpace(line), true
	}
	return "", false
}

func dashboardDeckLogPayload(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	switch {
	case strings.Contains(line, "### "):
		return strings.TrimSpace(line[strings.Index(line, "### "):])
	case strings.Contains(line, "# Deck ID: "):
		return strings.TrimSpace(line[strings.Index(line, "# Deck ID: "):])
	case isDashboardDeckCodeLine(line):
		fields := strings.Fields(line)
		return fields[len(fields)-1]
	case strings.Contains(line, "Deck Contents Received:"),
		strings.Contains(line, "Finding Game With Deck:"),
		strings.Contains(line, "Finished Editing Deck:"):
		return line
	}

	fields := strings.Fields(line)
	if len(fields) >= 3 && len(fields[0]) == 1 && strings.Contains(fields[1], ":") {
		return strings.Join(fields[2:], " ")
	}
	return line
}

func isDashboardDeckCodeLine(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}

	code := fields[len(fields)-1]
	return strings.HasPrefix(code, "AAE") && strings.Contains(code, "=")
}

func isDashboardDeckTriggerLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.Contains(line, "Deck Contents Received:") ||
		strings.Contains(line, "Finding Game With Deck:") ||
		strings.Contains(line, "Finished Editing Deck:")
}

func dashboardDeckIDFromCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	sum := sha1.Sum([]byte(code))
	return "deck-" + hex.EncodeToString(sum[:8])
}

func decodeDashboardDeckCards(deckCode string, db carddb.CardDatabase) ([]*card.Card, error) {
	dbfIDs, err := decodeDashboardDeckstring(deckCode)
	if err != nil {
		return nil, err
	}

	cards := make([]*card.Card, 0, len(dbfIDs))
	for _, dbfID := range dbfIDs {
		info, ok := db.GetCardByDBFID(dbfID)
		if !ok {
			return nil, fmt.Errorf("card with dbf id %d not found", dbfID)
		}

		c := card.New(info.ID, info.Cost, types.PlayerTypePlayer)
		c.Zone = card.ZoneDeck
		c.Info = info
		cards = append(cards, c)
	}

	return cards, nil
}

func decodeDashboardDeckstring(deckCode string) ([]int, error) {
	raw, err := base64.StdEncoding.DecodeString(deckCode)
	if err != nil {
		return nil, fmt.Errorf("base64 decode deckstring: %w", err)
	}

	reader := dashboardDeckstringReader{data: raw}

	header, err := reader.uvarint()
	if err != nil {
		return nil, err
	}
	if header != 0 {
		return nil, fmt.Errorf("unexpected deckstring header: %d", header)
	}

	if _, err := reader.uvarint(); err != nil {
		return nil, err
	}
	if _, err := reader.uvarint(); err != nil {
		return nil, err
	}

	heroCount, err := reader.uvarint()
	if err != nil {
		return nil, err
	}
	for i := uint64(0); i < heroCount; i++ {
		if _, err := reader.uvarint(); err != nil {
			return nil, err
		}
	}

	singles, err := readDashboardDeckGroup(&reader, 1)
	if err != nil {
		return nil, err
	}
	doubles, err := readDashboardDeckGroup(&reader, 2)
	if err != nil {
		return nil, err
	}
	multiples, err := readDashboardDeckMultiGroup(&reader)
	if err != nil {
		return nil, err
	}

	cards := make([]int, 0, len(singles)+len(doubles)+len(multiples))
	cards = append(cards, singles...)
	cards = append(cards, doubles...)
	cards = append(cards, multiples...)
	return cards, nil
}

func readDashboardDeckGroup(reader *dashboardDeckstringReader, copies int) ([]int, error) {
	count, err := reader.uvarint()
	if err != nil {
		return nil, err
	}

	result := make([]int, 0, count*uint64(copies))
	for i := uint64(0); i < count; i++ {
		dbfID, err := reader.uvarint()
		if err != nil {
			return nil, err
		}
		for j := 0; j < copies; j++ {
			result = append(result, int(dbfID))
		}
	}

	return result, nil
}

func readDashboardDeckMultiGroup(reader *dashboardDeckstringReader) ([]int, error) {
	count, err := reader.uvarint()
	if err != nil {
		return nil, err
	}

	result := make([]int, 0, count)
	for i := uint64(0); i < count; i++ {
		dbfID, err := reader.uvarint()
		if err != nil {
			return nil, err
		}
		copies, err := reader.uvarint()
		if err != nil {
			return nil, err
		}
		for j := uint64(0); j < copies; j++ {
			result = append(result, int(dbfID))
		}
	}

	return result, nil
}

type dashboardDeckstringReader struct {
	data []byte
	pos  int
}

func (r *dashboardDeckstringReader) uvarint() (uint64, error) {
	if r.pos >= len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}

	value, n := binary.Uvarint(r.data[r.pos:])
	if n <= 0 {
		return 0, fmt.Errorf("read deckstring varint at %d", r.pos)
	}
	r.pos += n
	return value, nil
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
