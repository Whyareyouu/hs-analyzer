package parser

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/types"
	"hsanalyzer/pkg/hslog"
)

// ZoneParser отслеживает перемещения карт между зонами и фиксирует добор.
type ZoneParser struct{}

type pendingDraw struct {
	entityID int
	cardID   string
	player   types.PlayerType
}

func NewZoneParser() *ZoneParser {
	return &ZoneParser{}
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
	var current *pendingDraw

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := scanner.Text()
		if strings.Contains(line, "PowerTaskList.DebugPrintPower()") {
			continue
		}

		if strings.Contains(line, "SHOW_ENTITY - Updating Entity=") {
			logged, ok, err := parseLoggedEntity(line)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", p.Name(), err)
			}
			if !ok || logged.zone != card.ZoneDeck {
				current = nil
				continue
			}

			matches := hslog.PowerShowEntityPattern.FindStringSubmatch(line)
			if matches == nil || matches[1] == "" {
				current = nil
				continue
			}

			current = &pendingDraw{
				entityID: logged.entityID,
				cardID:   matches[1],
				player:   logged.player,
			}
			continue
		}

		if current == nil {
			continue
		}

		if strings.Contains(line, "tag=ZONE value=HAND") {
			if current.player == types.PlayerTypeUnknown {
				current = nil
				continue
			}

			drawnCard := card.New(current.cardID, current.entityID, current.player)
			drawnCard.Zone = card.ZoneHand
			events = append(events, game.NewCardDrawnEvent(current.player, drawnCard))
			current = nil
			continue
		}

		if !strings.Contains(line, "tag=") && !strings.Contains(line, "SHOW_ENTITY - Updating Entity=") {
			current = nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: read zone log: %w", p.Name(), err)
	}

	return events, nil
}
