package parser

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/types"
)

// DeckParser извлекает выбранную колоду из Decks.log и связывает карты с игроком.
type DeckParser struct {
	cardDB            carddb.CardDatabase
	incrementalBuffer []string
	incrementalSeen   map[string]deckLogEntry
}

func NewDeckParser(cardDB carddb.CardDatabase) *DeckParser {
	return &DeckParser{cardDB: cardDB}
}

func (p *DeckParser) Name() string {
	return "deck_parser"
}

func (p *DeckParser) LogType() LogType {
	return LogTypeDeck
}

func (p *DeckParser) Parse(ctx context.Context, reader io.Reader) ([]game.Event, error) {
	if reader == nil {
		return nil, fmt.Errorf("%s: reader is nil", p.Name())
	}
	if p.cardDB == nil {
		return nil, fmt.Errorf("%s: card database is nil", p.Name())
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("%s: read deck log: %w", p.Name(), err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, selectedDeckID, err := parseDeckLogEntries(string(data))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", p.Name(), err)
	}
	if selectedDeckID == "" {
		return nil, nil
	}

	selected, ok := entries[selectedDeckID]
	if !ok {
		return nil, fmt.Errorf("selected deck %s not found in deck entries", selectedDeckID)
	}

	cards, err := p.decodeDeckCards(selected.Code)
	if err != nil {
		return nil, fmt.Errorf("%s: decode selected deck %s: %w", p.Name(), selectedDeckID, err)
	}

	return []game.Event{
		game.NewDeckListEvent(types.PlayerTypePlayer, cards, selected.Name, selected.DeckID, selected.Code),
	}, nil
}

type DeckCatalogEntry struct {
	Name     string
	DeckID   string
	Code     string
	Cards    []*card.Card
	Selected bool
}

// ParseCatalog extracts all known player decks from Decks.log while preserving
// stable order from the log.
func (p *DeckParser) ParseCatalog(ctx context.Context, reader io.Reader) ([]DeckCatalogEntry, string, error) {
	if reader == nil {
		return nil, "", fmt.Errorf("%s: reader is nil", p.Name())
	}
	if p.cardDB == nil {
		return nil, "", fmt.Errorf("%s: card database is nil", p.Name())
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("%s: read deck log: %w", p.Name(), err)
	}

	select {
	case <-ctx.Done():
		return nil, "", ctx.Err()
	default:
	}

	entries, selectedDeckID, err := parseDeckCatalog(string(data))
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", p.Name(), err)
	}

	result := make([]DeckCatalogEntry, 0, len(entries))
	for _, entry := range entries {
		cards, err := p.decodeDeckCards(entry.Code)
		if err != nil {
			return nil, "", fmt.Errorf("%s: decode deck %s: %w", p.Name(), entry.DeckID, err)
		}
		result = append(result, DeckCatalogEntry{
			Name:     entry.Name,
			DeckID:   entry.DeckID,
			Code:     entry.Code,
			Cards:    cards,
			Selected: entry.DeckID == selectedDeckID,
		})
	}

	return result, selectedDeckID, nil
}

// ParseLine parses one appended Decks.log line incrementally. It emits a
// DeckListEvent when a "Finding Game With Deck" block is completed.
func (p *DeckParser) ParseLine(line string) ([]game.Event, error) {
	if p == nil {
		return nil, fmt.Errorf("deck parser is nil")
	}
	if p.cardDB == nil {
		return nil, fmt.Errorf("%s: card database is nil", p.Name())
	}

	p.ensureIncrementalState()
	trimmed := strings.TrimSpace(line)

	if isDeckTriggerLine(trimmed) {
		events, err := p.flushIncrementalBlock()
		if err != nil {
			return nil, err
		}
		p.incrementalBuffer = []string{trimmed}
		return events, nil
	}

	if len(p.incrementalBuffer) == 0 {
		return nil, nil
	}

	p.incrementalBuffer = append(p.incrementalBuffer, trimmed)
	if trimmed == "" || isDeckCodeLine(trimmed) {
		return p.flushIncrementalBlock()
	}

	return nil, nil
}

func (p *DeckParser) ParseStream(ctx context.Context, reader io.Reader) ([]game.Event, error) {
	if reader == nil {
		return nil, fmt.Errorf("%s: reader is nil", p.Name())
	}
	if p.cardDB == nil {
		return nil, fmt.Errorf("%s: card database is nil", p.Name())
	}

	p.resetIncrementalState()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	events := make([]game.Event, 0)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		parsedEvents, err := p.ParseLine(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.Name(), err)
		}
		events = append(events, parsedEvents...)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: read deck log: %w", p.Name(), err)
	}

	parsedEvents, err := p.flushIncrementalBlock()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", p.Name(), err)
	}
	events = append(events, parsedEvents...)
	return events, nil
}

type deckLogEntry struct {
	Name   string
	DeckID string
	Code   string
}

func parseDeckLogEntries(raw string) (map[string]deckLogEntry, string, error) {
	catalog, selectedDeckID, err := parseDeckCatalog(raw)
	if err != nil {
		return nil, "", err
	}

	entries := make(map[string]deckLogEntry, len(catalog))
	for _, entry := range catalog {
		entries[entry.DeckID] = entry
	}
	return entries, selectedDeckID, nil
}

func parseDeckCatalog(raw string) ([]deckLogEntry, string, error) {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	entries := make(map[string]deckLogEntry)
	order := make([]string, 0)
	selectedDeckID := ""
	currentTrigger := ""

	for i := 0; i < len(lines); i++ {
		line := deckLogPayload(lines[i])
		if isDeckTriggerLine(line) {
			currentTrigger = line
			continue
		}

		if _, ok := parseDeckNameLine(line, currentTrigger != ""); !ok {
			continue
		}

		entry, consumed, err := parseDeckBlock(lines, i)
		if err != nil {
			return nil, "", err
		}
		if entry.DeckID != "" {
			if _, exists := entries[entry.DeckID]; !exists {
				order = append(order, entry.DeckID)
			}
			entries[entry.DeckID] = entry
		}
		if strings.Contains(currentTrigger, "Finding Game With Deck:") {
			selectedDeckID = entry.DeckID
			currentTrigger = ""
		}
		i = consumed
	}

	if selectedDeckID == "" && len(order) > 0 {
		selectedDeckID = order[len(order)-1]
	}

	result := make([]deckLogEntry, 0, len(order))
	for _, deckID := range order {
		result = append(result, entries[deckID])
	}
	return result, selectedDeckID, nil
}

func parseDeckBlock(lines []string, start int) (deckLogEntry, int, error) {
	entry := deckLogEntry{}
	i := start
	for ; i < len(lines); i++ {
		line := deckLogPayload(lines[i])
		if line == "" {
			break
		}
		if isDeckTriggerLine(line) {
			i--
			break
		}
		if entry.Name != "" {
			if _, ok := parseDeckNameLine(line, true); ok && entry.Code == "" {
				i--
				break
			}
		}
		if entry.Name == "" {
			name, ok := parseDeckNameLine(line, true)
			if ok {
				entry.Name = name
				continue
			}
		}
		if entry.DeckID == "" && strings.Contains(line, "# Deck ID: ") {
			entry.DeckID = strings.TrimSpace(strings.SplitN(line, "# Deck ID: ", 2)[1])
			continue
		}
		if entry.Code == "" && isDeckCodeLine(line) {
			fields := strings.Fields(line)
			entry.Code = fields[len(fields)-1]
			if entry.DeckID == "" {
				entry.DeckID = deckIDFromCode(entry.Code)
			}
			break
		}
	}

	if entry.Name == "" && entry.DeckID != "" {
		entry.Name = entry.DeckID
	}
	if entry.DeckID == "" || entry.Code == "" {
		return deckLogEntry{}, i, fmt.Errorf("incomplete deck block: %#v", entry)
	}

	return entry, i, nil
}

func parseDeckNameLine(line string, allowPlain bool) (string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || isDeckTriggerLine(line) || strings.Contains(line, "# Deck ID: ") || isDeckCodeLine(line) {
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

func deckLogPayload(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	switch {
	case strings.Contains(line, "### "):
		return strings.TrimSpace(line[strings.Index(line, "### "):])
	case strings.Contains(line, "# Deck ID: "):
		return strings.TrimSpace(line[strings.Index(line, "# Deck ID: "):])
	case isDeckCodeLine(line):
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

func deckIDFromCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	sum := sha1.Sum([]byte(code))
	return "deck-" + hex.EncodeToString(sum[:8])
}

func isDeckCodeLine(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}

	code := fields[len(fields)-1]
	return strings.HasPrefix(code, "AAE") && strings.Contains(code, "=")
}

func isDeckTriggerLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.Contains(line, "Deck Contents Received:") ||
		strings.Contains(line, "Finding Game With Deck:") ||
		strings.Contains(line, "Finished Editing Deck:")
}

func (p *DeckParser) ensureIncrementalState() {
	if p.incrementalSeen == nil {
		p.incrementalSeen = make(map[string]deckLogEntry)
	}
}

func (p *DeckParser) resetIncrementalState() {
	p.incrementalBuffer = nil
	p.incrementalSeen = make(map[string]deckLogEntry)
}

func (p *DeckParser) flushIncrementalBlock() ([]game.Event, error) {
	if len(p.incrementalBuffer) == 0 {
		return nil, nil
	}

	lines := append([]string(nil), p.incrementalBuffer...)
	p.incrementalBuffer = nil
	trigger := strings.TrimSpace(lines[0])
	entry, _, err := parseDeckBlock(lines, 1)
	if err != nil {
		return nil, err
	}
	if entry.DeckID == "" {
		return nil, nil
	}
	p.incrementalSeen[entry.DeckID] = entry

	if !strings.Contains(trigger, "Finding Game With Deck:") {
		return nil, nil
	}

	cards, err := p.decodeDeckCards(entry.Code)
	if err != nil {
		return nil, fmt.Errorf("decode selected deck %s: %w", entry.DeckID, err)
	}
	return []game.Event{
		game.NewDeckListEvent(types.PlayerTypePlayer, cards, entry.Name, entry.DeckID, entry.Code),
	}, nil
}

func (p *DeckParser) decodeDeckCards(deckCode string) ([]*card.Card, error) {
	dbfIDs, err := decodeDeckstring(deckCode)
	if err != nil {
		return nil, err
	}

	cards := make([]*card.Card, 0, len(dbfIDs))
	for _, dbfID := range dbfIDs {
		info, ok := p.cardDB.GetCardByDBFID(dbfID)
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

func decodeDeckstring(deckCode string) ([]int, error) {
	raw, err := base64.StdEncoding.DecodeString(deckCode)
	if err != nil {
		return nil, fmt.Errorf("base64 decode deckstring: %w", err)
	}

	reader := deckstringReader{data: raw}

	header, err := reader.uvarint()
	if err != nil {
		return nil, err
	}
	if header != 0 {
		return nil, fmt.Errorf("unexpected deckstring header: %d", header)
	}

	if _, err := reader.uvarint(); err != nil { // version
		return nil, err
	}
	if _, err := reader.uvarint(); err != nil { // format
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

	singles, err := readDeckGroup(&reader, 1)
	if err != nil {
		return nil, err
	}
	doubles, err := readDeckGroup(&reader, 2)
	if err != nil {
		return nil, err
	}
	multiples, err := readDeckMultiGroup(&reader)
	if err != nil {
		return nil, err
	}

	cards := make([]int, 0, len(singles)+len(doubles)+len(multiples))
	cards = append(cards, singles...)
	cards = append(cards, doubles...)
	cards = append(cards, multiples...)
	return cards, nil
}

func readDeckGroup(reader *deckstringReader, copies int) ([]int, error) {
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

func readDeckMultiGroup(reader *deckstringReader) ([]int, error) {
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

type deckstringReader struct {
	data []byte
	pos  int
}

func (r *deckstringReader) uvarint() (uint64, error) {
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
