package parser

import (
	"context"
	"encoding/base64"
	"encoding/binary"
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
	cardDB carddb.CardDatabase
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

type deckLogEntry struct {
	Name   string
	DeckID string
	Code   string
}

func parseDeckLogEntries(raw string) (map[string]deckLogEntry, string, error) {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	entries := make(map[string]deckLogEntry)
	selectedDeckID := ""

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		switch {
		case strings.Contains(line, "Deck Contents Received:"),
			strings.Contains(line, "Finding Game With Deck:"),
			strings.Contains(line, "Finished Editing Deck:"):
			entry, consumed, err := parseDeckBlock(lines, i+1)
			if err != nil {
				return nil, "", err
			}
			if entry.DeckID != "" {
				entries[entry.DeckID] = entry
			}
			if strings.Contains(line, "Finding Game With Deck:") {
				selectedDeckID = entry.DeckID
			}
			i = consumed
		}
	}

	return entries, selectedDeckID, nil
}

func parseDeckBlock(lines []string, start int) (deckLogEntry, int, error) {
	entry := deckLogEntry{}
	i := start
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			break
		}
		if strings.Contains(line, "Deck Contents Received:") ||
			strings.Contains(line, "Finding Game With Deck:") ||
			strings.Contains(line, "Finished Editing Deck:") {
			i--
			break
		}
		if entry.Name == "" && strings.Contains(line, "### ") {
			entry.Name = strings.TrimSpace(strings.SplitN(line, "### ", 2)[1])
			continue
		}
		if entry.DeckID == "" && strings.Contains(line, "# Deck ID: ") {
			entry.DeckID = strings.TrimSpace(strings.SplitN(line, "# Deck ID: ", 2)[1])
			continue
		}
		if entry.Code == "" && isDeckCodeLine(line) {
			fields := strings.Fields(line)
			entry.Code = fields[len(fields)-1]
			break
		}
	}

	if entry.DeckID == "" || entry.Code == "" {
		return deckLogEntry{}, i, fmt.Errorf("incomplete deck block: %#v", entry)
	}

	return entry, i, nil
}

func isDeckCodeLine(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}

	code := fields[len(fields)-1]
	return strings.HasPrefix(code, "AAE") && strings.Contains(code, "=")
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

		c := card.New(info.ID, 0, types.PlayerTypePlayer)
		c.Zone = card.ZoneDeck
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
