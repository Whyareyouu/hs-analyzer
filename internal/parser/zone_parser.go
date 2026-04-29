package parser

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/types"
	"hsanalyzer/pkg/hslog"
)

// ZoneParser отслеживает перемещения карт между зонами и фиксирует добор.
type ZoneParser struct {
	currentEntityID   int
	localPlayerNumber int
	pendingDraws      map[int]*pendingDraw
	emittedDraws      map[int]struct{}
}

type pendingDraw struct {
	entityID int
	cardID   string
	player   types.PlayerType
	moved    bool
}

func NewZoneParser() *ZoneParser {
	return &ZoneParser{
		localPlayerNumber: 1,
		pendingDraws:      make(map[int]*pendingDraw),
		emittedDraws:      make(map[int]struct{}),
	}
}

func (p *ZoneParser) Name() string {
	return "zone_parser"
}

func (p *ZoneParser) LogType() LogType {
	return LogTypeZone
}

func (p *ZoneParser) Parse(ctx context.Context, reader io.Reader) ([]game.Event, error) {
	if reader == nil {
		return nil, fmt.Errorf("%s: reader is nil", p.Name())
	}

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	events := make([]game.Event, 0)
	p.currentEntityID = 0
	p.localPlayerNumber = 1
	p.pendingDraws = make(map[int]*pendingDraw)
	p.emittedDraws = make(map[int]struct{})

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
		return nil, fmt.Errorf("%s: read zone log: %w", p.Name(), err)
	}

	return events, nil
}

// ParseLine parses one appended zone line incrementally.
func (p *ZoneParser) ParseLine(line string) ([]game.Event, error) {
	if p == nil {
		return nil, fmt.Errorf("zone parser is nil")
	}
	p.detectLocalPlayerNumber(line)

	if strings.Contains(line, "PowerTaskList.DebugPrintPower()") {
		return nil, nil
	}

	if strings.Contains(line, "SHOW_ENTITY - Updating Entity=") {
		return p.parseShowEntityLine(line)
	}

	if strings.Contains(line, "TAG_CHANGE Entity=") && strings.Contains(line, "tag=ZONE value=HAND") {
		return p.parseZoneToHandTag(line)
	}

	if strings.Contains(line, "tag=ZONE value=HAND") {
		events, emitted := p.emitCurrentEntityDraw()
		if emitted {
			return events, nil
		}
	}

	if !strings.Contains(line, "tag=") && !strings.Contains(line, "SHOW_ENTITY - Updating Entity=") {
		p.currentEntityID = 0
	}
	return nil, nil
}

func (p *ZoneParser) parseShowEntityLine(line string) ([]game.Event, error) {
	logged, ok, err := p.parseLoggedEntity(line)
	if err != nil {
		return nil, err
	}
	if !ok {
		p.currentEntityID = 0
		return nil, nil
	}

	matches := hslog.PowerShowEntityPattern.FindStringSubmatch(line)
	cardID := ""
	if matches != nil {
		cardID = matches[1]
	}

	switch {
	case logged.zone == card.ZoneDeck && cardID != "":
		p.currentEntityID = logged.entityID
		p.trackPendingDraw(logged.entityID, cardID, logged.player)
		return nil, nil
	case logged.zone == card.ZoneHand && cardID != "":
		p.currentEntityID = 0
		return p.resolvePendingDraw(logged.entityID, cardID, logged.player), nil
	default:
		p.currentEntityID = 0
		return nil, nil
	}
}

func (p *ZoneParser) parseZoneToHandTag(line string) ([]game.Event, error) {
	logged, ok, err := p.parseLoggedEntity(line)
	if err != nil {
		return nil, err
	}
	if !ok || logged.zone != card.ZoneDeck {
		return nil, nil
	}

	p.currentEntityID = logged.entityID
	p.trackPendingDraw(logged.entityID, logged.cardID, logged.player)
	p.markPendingMoved(logged.entityID)

	if logged.cardID != "" {
		if events, emitted := p.emitDraw(logged.entityID); emitted {
			return events, nil
		}
	}
	return nil, nil
}

func (p *ZoneParser) trackPendingDraw(entityID int, cardID string, player types.PlayerType) {
	if entityID <= 0 {
		return
	}
	current, ok := p.pendingDraws[entityID]
	if !ok {
		current = &pendingDraw{
			entityID: entityID,
			player:   player,
		}
		p.pendingDraws[entityID] = current
	}
	if strings.TrimSpace(cardID) != "" {
		current.cardID = cardID
	}
	if player != types.PlayerTypeUnknown {
		current.player = player
	}
}

func (p *ZoneParser) markPendingMoved(entityID int) {
	if entityID <= 0 {
		return
	}
	current, ok := p.pendingDraws[entityID]
	if !ok {
		current = &pendingDraw{entityID: entityID}
		p.pendingDraws[entityID] = current
	}
	current.moved = true
}

func (p *ZoneParser) resolvePendingDraw(entityID int, cardID string, player types.PlayerType) []game.Event {
	if entityID <= 0 {
		return nil
	}
	p.trackPendingDraw(entityID, cardID, player)
	if events, emitted := p.emitDraw(entityID); emitted {
		return events
	}
	return nil
}

func (p *ZoneParser) emitCurrentEntityDraw() ([]game.Event, bool) {
	if p.currentEntityID == 0 {
		return nil, false
	}
	p.markPendingMoved(p.currentEntityID)
	events, emitted := p.emitDraw(p.currentEntityID)
	p.currentEntityID = 0
	return events, emitted
}

func (p *ZoneParser) emitDraw(entityID int) ([]game.Event, bool) {
	if _, exists := p.emittedDraws[entityID]; exists {
		return nil, false
	}

	current, ok := p.pendingDraws[entityID]
	if !ok || current == nil || !current.moved || current.player == types.PlayerTypeUnknown || strings.TrimSpace(current.cardID) == "" {
		return nil, false
	}

	drawnCard := card.New(current.cardID, current.entityID, current.player)
	drawnCard.Zone = card.ZoneHand
	p.emittedDraws[entityID] = struct{}{}
	delete(p.pendingDraws, entityID)
	return []game.Event{game.NewCardDrawnEvent(current.player, drawnCard)}, true
}

func (p *ZoneParser) parseLoggedEntity(line string) (loggedEntity, bool, error) {
	matches := hslog.PowerEntityPattern.FindStringSubmatch(line)
	if matches == nil {
		return loggedEntity{}, false, nil
	}

	entityID, err := strconv.Atoi(matches[2])
	if err != nil {
		return loggedEntity{}, false, fmt.Errorf("invalid entity id %q", matches[2])
	}
	zonePos, err := strconv.Atoi(matches[4])
	if err != nil {
		return loggedEntity{}, false, fmt.Errorf("invalid zone position %q", matches[4])
	}
	playerNumber, err := strconv.Atoi(matches[6])
	if err != nil {
		return loggedEntity{}, false, fmt.Errorf("invalid player number %q", matches[6])
	}

	return loggedEntity{
		entityID: entityID,
		name:     matches[1],
		zone:     card.ZoneType(matches[3]),
		zonePos:  zonePos,
		cardID:   matches[5],
		player:   p.mapPlayerNumber(playerNumber),
	}, true, nil
}

func (p *ZoneParser) mapPlayerNumber(playerNumber int) types.PlayerType {
	if playerNumber <= 0 {
		return types.PlayerTypeUnknown
	}
	localPlayerNumber := p.localPlayerNumber
	if localPlayerNumber <= 0 {
		localPlayerNumber = 1
	}
	switch {
	case playerNumber == localPlayerNumber:
		return types.PlayerTypePlayer
	case playerNumber == 1 || playerNumber == 2:
		return types.PlayerTypeOpponent
	default:
		return types.PlayerTypeUnknown
	}
}

func (p *ZoneParser) detectLocalPlayerNumber(line string) {
	if p == nil {
		return
	}
	const marker = "PlayerName=UNKNOWN HUMAN PLAYER"
	if !strings.Contains(line, marker) {
		return
	}

	prefix := line[:strings.Index(line, marker)]
	lastEquals := strings.LastIndex(prefix, "=")
	if lastEquals < 0 {
		return
	}
	rawValue := strings.Trim(strings.TrimSpace(prefix[lastEquals+1:]), ",")
	playerNumber, err := strconv.Atoi(rawValue)
	if err != nil || playerNumber <= 0 {
		return
	}
	p.localPlayerNumber = playerNumber
}
