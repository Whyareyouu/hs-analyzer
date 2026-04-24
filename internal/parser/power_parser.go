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

// PowerParser содержит базовую структуру для парсинга Power.log.
type PowerParser struct {
	currentTurnNumber int
	currentTurnPlayer types.PlayerType
	hasActiveTurn     bool
	pendingPlays      map[int]pendingPlay
	emittedPlays      map[int]struct{}
	damageTotals      map[int]int
}

type pendingPlay struct {
	entityID int
	player   types.PlayerType
	cardID   string
}

func NewPowerParser() *PowerParser {
	return &PowerParser{}
}

func (p *PowerParser) Name() string {
	return "power_parser"
}

func (p *PowerParser) LogType() LogType {
	return LogTypePower
}

func (p *PowerParser) Parse(ctx context.Context, reader io.Reader) ([]game.Event, error) {
	if reader == nil {
		return nil, fmt.Errorf("%s: reader is nil", p.Name())
	}

	p.reset()

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	events := make([]game.Event, 0)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := scanner.Text()
		if !p.isRelevantLine(line) {
			continue
		}

		parsedEvents, err := p.parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.Name(), err)
		}
		events = append(events, parsedEvents...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: read power log: %w", p.Name(), err)
	}

	return events, nil
}

func (p *PowerParser) isRelevantLine(line string) bool {
	return hslog.PowerGameTypePattern.MatchString(line) ||
		hslog.PowerTagChangePattern.MatchString(line) ||
		hslog.PowerFullEntityPattern.MatchString(line) ||
		hslog.PowerShowEntityPattern.MatchString(line) ||
		hslog.PowerBlockStartPattern.MatchString(line)
}

func (p *PowerParser) parseLine(line string) ([]game.Event, error) {
	if matches := hslog.PowerBlockStartPattern.FindStringSubmatch(line); matches != nil {
		return p.parseBlockStart(line, matches[1])
	}

	if matches := hslog.PowerTagChangePattern.FindStringSubmatch(line); matches != nil {
		if strings.Contains(line, "PowerTaskList.DebugPrintPower()") {
			return nil, nil
		}
		return p.parseTagChange(matches[1], matches[2], matches[3])
	}

	return nil, nil
}

func (p *PowerParser) reset() {
	p.currentTurnNumber = 0
	p.currentTurnPlayer = types.PlayerTypeUnknown
	p.hasActiveTurn = false
	p.pendingPlays = make(map[int]pendingPlay)
	p.emittedPlays = make(map[int]struct{})
	p.damageTotals = make(map[int]int)
}

func (p *PowerParser) parseTagChange(entity, tag, value string) ([]game.Event, error) {
	switch tag {
	case hslog.TagTurn:
		turnNumber, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid turn value %q", value)
		}

		playerType := identifyTurnPlayer(entity)
		if playerType == types.PlayerTypeUnknown {
			return nil, nil
		}

		if p.hasActiveTurn && p.currentTurnNumber == turnNumber && p.currentTurnPlayer == playerType {
			return nil, nil
		}

		events := make([]game.Event, 0, 2)
		if p.hasActiveTurn {
			events = append(events, game.NewTurnEndEvent())
		}

		events = append(events, game.NewTurnStartEvent(playerType))
		p.currentTurnNumber = turnNumber
		p.currentTurnPlayer = playerType
		p.hasActiveTurn = true

		return events, nil
	case "DAMAGE":
		logged, ok, err := parseLoggedEntityFromRawEntity(entity)
		if err != nil || !ok {
			return nil, err
		}

		damageTotal, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid damage value %q", value)
		}

		previous := p.damageTotals[logged.entityID]
		p.damageTotals[logged.entityID] = damageTotal
		if damageTotal <= previous {
			return nil, nil
		}

		return []game.Event{
			game.NewDamageEvent(logged.entityID, damageTotal-previous),
		}, nil
	case hslog.TagZone:
		if value != string(card.ZoneGraveyard) {
			return nil, nil
		}

		logged, ok, err := parseLoggedEntityFromRawEntity(entity)
		if err != nil || !ok {
			return nil, err
		}

		return []game.Event{
			game.NewDeathEvent(logged.entityID),
		}, nil
	default:
		return nil, nil
	}
}

func identifyTurnPlayer(entity string) types.PlayerType {
	trimmed := strings.TrimSpace(entity)

	switch trimmed {
	case "UNKNOWN HUMAN PLAYER", "1":
		return types.PlayerTypePlayer
	case "2", "GameEntity":
		return types.PlayerTypeUnknown
	}

	if strings.Contains(trimmed, "#") {
		return types.PlayerTypeOpponent
	}

	return types.PlayerTypeUnknown
}

func (p *PowerParser) parseBlockStart(line, blockType string) ([]game.Event, error) {
	entity, ok, err := parseLoggedEntity(line)
	if err != nil || !ok {
		return nil, err
	}

	switch blockType {
	case hslog.BlockTypeAttack:
		if strings.Contains(line, "PowerTaskList.DebugPrintPower()") {
			return nil, nil
		}

		target, ok, err := parseLoggedTarget(line)
		if err != nil || !ok {
			return nil, err
		}

		return []game.Event{
			game.NewAttackEvent(entity.entityID, target.entityID),
		}, nil
	case hslog.BlockTypePlay:
		p.pendingPlays[entity.entityID] = pendingPlay{
			entityID: entity.entityID,
			player:   entity.player,
			cardID:   entity.cardID,
		}
		return nil, nil
	case hslog.BlockTypePower:
		pending, exists := p.pendingPlays[entity.entityID]
		if !exists || entity.zone != card.ZonePlay {
			return nil, nil
		}
		if _, emitted := p.emittedPlays[entity.entityID]; emitted {
			return nil, nil
		}

		cardID := entity.cardID
		if cardID == "" {
			cardID = pending.cardID
		}
		if cardID == "" {
			return nil, nil
		}

		playedCard := card.New(cardID, entity.entityID, pending.player)
		playedCard.Zone = entity.zone
		playedCard.Position = entity.zonePos

		p.emittedPlays[entity.entityID] = struct{}{}
		delete(p.pendingPlays, entity.entityID)

		return []game.Event{
			game.NewCardPlayedEvent(pending.player, playedCard, entity.zonePos),
		}, nil
	default:
		return nil, nil
	}
}

type loggedEntity struct {
	entityID int
	zone     card.ZoneType
	zonePos  int
	cardID   string
	player   types.PlayerType
}

func parseLoggedEntity(line string) (loggedEntity, bool, error) {
	matches := hslog.PowerEntityPattern.FindStringSubmatch(line)
	if matches == nil {
		return loggedEntity{}, false, nil
	}

	return buildLoggedEntity(matches)
}

func parseLoggedEntityFromRawEntity(raw string) (loggedEntity, bool, error) {
	matches := hslog.PowerEntityPattern.FindStringSubmatch("Entity=" + raw)
	if matches == nil {
		return loggedEntity{}, false, nil
	}

	return buildLoggedEntity(matches)
}

func parseLoggedTarget(line string) (loggedEntity, bool, error) {
	matches := hslog.PowerTargetPattern.FindStringSubmatch(line)
	if matches == nil {
		return loggedEntity{}, false, nil
	}

	return buildLoggedEntity(matches)
}

func buildLoggedEntity(matches []string) (loggedEntity, bool, error) {
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
		zone:     card.ZoneType(matches[3]),
		zonePos:  zonePos,
		cardID:   matches[5],
		player:   mapPlayerNumber(playerNumber),
	}, true, nil
}

func mapPlayerNumber(playerNumber int) types.PlayerType {
	switch playerNumber {
	case 1:
		return types.PlayerTypePlayer
	case 2:
		return types.PlayerTypeOpponent
	default:
		return types.PlayerTypeUnknown
	}
}
