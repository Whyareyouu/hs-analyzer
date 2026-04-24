package parser

import (
	"context"
	"strings"
	"testing"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/types"
)

func TestZoneParserParseCardDrawnFromRealPowerLogFragment(t *testing.T) {
	t.Parallel()

	input := findLogWindow(t, "Power.log", "SHOW_ENTITY - Updating Entity=[entityName=UNKNOWN ENTITY [cardType=INVALID] id=39 zone=DECK zonePos=0 cardId= player=2] CardID=EDR_528", 12)

	parser := NewZoneParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse zone fragment: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	drawEvent, ok := events[0].(*game.CardDrawnEvent)
	if !ok {
		t.Fatalf("unexpected event type: %T", events[0])
	}
	if drawEvent.Player != types.PlayerTypeOpponent {
		t.Fatalf("unexpected player: %s", drawEvent.Player)
	}
	if drawEvent.Card == nil {
		t.Fatalf("expected drawn card")
	}
	if drawEvent.Card.ID != "EDR_528" {
		t.Fatalf("unexpected card id: %s", drawEvent.Card.ID)
	}
	if drawEvent.Card.EntityID != 39 {
		t.Fatalf("unexpected entity id: %d", drawEvent.Card.EntityID)
	}
}

func TestZoneParserMetadata(t *testing.T) {
	t.Parallel()

	parser := NewZoneParser()
	if parser.Name() != "zone_parser" {
		t.Fatalf("unexpected parser name: %s", parser.Name())
	}
	if parser.LogType() != LogTypeZone {
		t.Fatalf("unexpected log type: %s", parser.LogType())
	}
}
