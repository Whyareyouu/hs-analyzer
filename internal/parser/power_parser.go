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
	statValues        map[int]map[string]int
	discoverChoices   map[int]discoverChoice
	activeChoiceID    int
	collectChoiceID   int
}

type pendingPlay struct {
	entityID int
	player   types.PlayerType
	cardID   string
}

type discoverChoice struct {
	player  types.PlayerType
	options []string
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
		hslog.PowerBlockStartPattern.MatchString(line) ||
		hslog.PowerEntityChoicesHeaderPattern.MatchString(line) ||
		hslog.PowerChoiceEntityPattern.MatchString(line) ||
		hslog.PowerSendChoicesPattern.MatchString(line) ||
		hslog.PowerChosenEntityPattern.MatchString(line)
}

func (p *PowerParser) parseLine(line string) ([]game.Event, error) {
	if strings.Contains(line, "m_chosenEntities[0]=") {
		return p.parseChosenEntity(line)
	}
	if strings.Contains(line, "Entities[") && strings.Contains(line, "DebugPrintEntityChoices()") {
		return p.parseChoiceOption(line)
	}
	if strings.Contains(line, "GameState.DebugPrintEntityChoices()") {
		return p.parseEntityChoicesHeader(line)
	}
	if strings.Contains(line, "GameState.SendChoices()") {
		return p.parseSendChoices(line)
	}

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
	p.statValues = make(map[int]map[string]int)
	p.discoverChoices = make(map[int]discoverChoice)
	p.activeChoiceID = 0
	p.collectChoiceID = 0
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
	case "ATK", "HEALTH":
		logged, ok, err := parseLoggedEntityFromRawEntity(entity)
		if err != nil || !ok {
			return nil, err
		}

		statValue, err := strconv.Atoi(value)
		if err != nil {
			return nil, nil
		}

		if _, ok := p.statValues[logged.entityID]; !ok {
			p.statValues[logged.entityID] = make(map[string]int)
		}
		p.statValues[logged.entityID][tag] = statValue

		return []game.Event{
			game.NewEffectTriggeredEvent(logged.entityID, fmt.Sprintf("%s_CHANGE:%d", tag, statValue)),
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

func (p *PowerParser) parseEntityChoicesHeader(line string) ([]game.Event, error) {
	if strings.Contains(line, "PowerTaskList.DebugPrintPower()") {
		return nil, nil
	}

	matches := hslog.PowerEntityChoicesHeaderPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, nil
	}

	choiceID, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid choice id %q", matches[1])
	}
	if matches[3] != "GENERAL" {
		return nil, nil
	}

	player := identifyTurnPlayer(matches[2])
	if player == types.PlayerTypeUnknown {
		return nil, nil
	}

	p.discoverChoices[choiceID] = discoverChoice{
		player:  player,
		options: make([]string, 0, 3),
	}
	p.collectChoiceID = choiceID
	return nil, nil
}

func (p *PowerParser) parseChoiceOption(line string) ([]game.Event, error) {
	matches := hslog.PowerChoiceEntityPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, nil
	}

	entity, ok, err := buildLoggedEntity(matches)
	if err != nil || !ok {
		return nil, err
	}
	if entity.zone != card.ZoneSetAside || entity.cardID == "" {
		return nil, nil
	}

	if p.collectChoiceID == 0 {
		return nil, nil
	}

	choice, ok := p.discoverChoices[p.collectChoiceID]
	if !ok {
		return nil, nil
	}
	choice.options = append(choice.options, entity.cardID)
	p.discoverChoices[p.collectChoiceID] = choice

	return nil, nil
}

func (p *PowerParser) parseSendChoices(line string) ([]game.Event, error) {
	matches := hslog.PowerSendChoicesPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, nil
	}
	if matches[2] != "GENERAL" {
		return nil, nil
	}

	choiceID, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid send choice id %q", matches[1])
	}
	p.activeChoiceID = choiceID
	return nil, nil
}

func (p *PowerParser) parseChosenEntity(line string) ([]game.Event, error) {
	if p.activeChoiceID == 0 {
		return nil, nil
	}

	choice, ok := p.discoverChoices[p.activeChoiceID]
	if !ok || len(choice.options) < 2 {
		return nil, nil
	}

	matches := hslog.PowerChosenEntityPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, nil
	}

	chosen, ok, err := buildLoggedEntity(matches)
	if err != nil || !ok {
		return nil, err
	}
	if chosen.cardID == "" {
		return nil, nil
	}

	event := game.NewDiscoverEvent(choice.player, append([]string(nil), choice.options...), chosen.cardID, chosen.entityID)
	delete(p.discoverChoices, p.activeChoiceID)
	p.activeChoiceID = 0
	p.collectChoiceID = 0

	return []game.Event{event}, nil
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
	case "TRIGGER":
		if strings.Contains(line, "PowerTaskList.DebugPrintPower()") {
			return nil, nil
		}

		keywordMatch := hslog.PowerTriggerKeywordPattern.FindStringSubmatch(line)
		if keywordMatch == nil {
			return nil, nil
		}
		keyword := keywordMatch[1]
		if keyword == "TAG_NOT_SET" {
			return nil, nil
		}

		sourceEntityID := 0
		if ok {
			sourceEntityID = entity.entityID
		}

		return []game.Event{
			game.NewEffectTriggeredEvent(sourceEntityID, keyword),
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
