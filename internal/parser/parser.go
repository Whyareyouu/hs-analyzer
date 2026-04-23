// internal/parser/parser.go
package parser

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"hs-analyzer/internal/domain"
)

type Parser struct {
	reTimestamp  *regexp.Regexp
	reCreateGame *regexp.Regexp
	rePlayer     *regexp.Regexp
	reFullEntity *regexp.Regexp
	reTagChange  *regexp.Regexp
	reBlockStart *regexp.Regexp
	reBlockEnd   *regexp.Regexp
	reShowEntity *regexp.Regexp
}

func New() *Parser {
	return &Parser{
		reTimestamp:  regexp.MustCompile(`^D\s+(\d{2}:\d{2}:\d{2}\.\d+)`),
		reCreateGame: regexp.MustCompile(`CREATE_GAME`),
		rePlayer:     regexp.MustCompile(`Player EntityID=(\d+) PlayerID=(\d+)`),
		reFullEntity: regexp.MustCompile(`FULL_ENTITY - Creating ID=(\d+)(?: CardID=(\S+))?`),
		reTagChange:  regexp.MustCompile(`TAG_CHANGE Entity=(?:\[.*?id=)?(?:(\d+)|([A-Za-z]+))(?:\])?.*?tag=(\w+)\s+value=(\S+)`),
		reBlockStart: regexp.MustCompile(`BLOCK_START BlockType=(\w+)\s+Entity=(?:\[.*?id=)?(\d+)(?:.*?cardId=(\S+))?(?:.*?Target=(?:\[.*?id=)?(\d+))?`),
		reBlockEnd:   regexp.MustCompile(`BLOCK_END`),
		reShowEntity: regexp.MustCompile(`SHOW_ENTITY - Updating Entity=(\d+)\s+CardID=(\S+)`),
	}
}

func (p *Parser) ParseLines(lines []string) []domain.RawEvent {
	events := make([]domain.RawEvent, 0)
	for _, line := range lines {
		if event, ok := p.ParseLine(line); ok {
			events = append(events, event)
		}
	}
	return events
}

func (p *Parser) ParseLine(line string) (domain.RawEvent, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || !strings.Contains(trimmed, "DebugPrintPower()") {
		return domain.RawEvent{}, false
	}

	timestamp := p.extractTimestamp(trimmed)
	eventType, data := p.parseEvent(trimmed)
	
	if eventType == "" {
		return domain.RawEvent{}, false
	}

	return domain.RawEvent{
		Timestamp: timestamp,
		Type:      eventType,
		Data:      data,
		RawLine:   line,
	}, true
}

func (p *Parser) extractTimestamp(line string) time.Time {
	matches := p.reTimestamp.FindStringSubmatch(line)
	if len(matches) < 2 {
		return time.Now()
	}

	timeStr := matches[1]
	today := time.Now().Format("2006-01-02")
	fullTimeStr := today + " " + timeStr

	ts, _ := time.Parse("2006-01-02 15:04:05.0000000", fullTimeStr)
	return ts
}

func (p *Parser) parseEvent(line string) (string, map[string]interface{}) {
	data := make(map[string]interface{})

	if p.reCreateGame.MatchString(line) {
		return "CREATE_GAME", data
	}

	if matches := p.rePlayer.FindStringSubmatch(line); matches != nil {
		entityID, _ := strconv.Atoi(matches[1])
		playerID, _ := strconv.Atoi(matches[2])
		data["entity_id"] = entityID
		data["player_id"] = playerID
		return "PLAYER", data
	}

	if matches := p.reFullEntity.FindStringSubmatch(line); matches != nil {
		entityID, _ := strconv.Atoi(matches[1])
		data["entity_id"] = entityID
		if len(matches) > 2 && matches[2] != "" {
			data["card_id"] = matches[2]
		}
		return "FULL_ENTITY", data
	}

	if matches := p.reTagChange.FindStringSubmatch(line); matches != nil {
		if matches[1] != "" {
			entityID, _ := strconv.Atoi(matches[1])
			data["entity_id"] = entityID
		} else if matches[2] != "" {
			data["entity_name"] = matches[2]
		}
		data["tag"] = matches[3]
		data["value"] = matches[4]
		return "TAG_CHANGE", data
	}

	if matches := p.reBlockStart.FindStringSubmatch(line); matches != nil {
		data["block_type"] = matches[1]
		if matches[2] != "" {
			entityID, _ := strconv.Atoi(matches[2])
			data["entity_id"] = entityID
		}
		if len(matches) > 3 && matches[3] != "" {
			data["card_id"] = matches[3]
		}
		if len(matches) > 4 && matches[4] != "" && matches[4] != "0" {
			targetID, _ := strconv.Atoi(matches[4])
			data["target_id"] = targetID
		}
		return "BLOCK_START", data
	}

	if p.reBlockEnd.MatchString(line) {
		return "BLOCK_END", data
	}

	if matches := p.reShowEntity.FindStringSubmatch(line); matches != nil {
		entityID, _ := strconv.Atoi(matches[1])
		data["entity_id"] = entityID
		data["card_id"] = matches[2]
		return "SHOW_ENTITY", data
	}

	return "", data
}