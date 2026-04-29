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

func TestZoneParserParseLineSupportsIncrementalStreamAndLocalPlayerMapping(t *testing.T) {
	t.Parallel()

	parser := NewZoneParser()
	lines := []string{
		"D 00:00:00.0000000 GameState.DebugPrintGame() - PlayerID=2, PlayerName=UNKNOWN HUMAN PLAYER",
		"D 00:00:00.0000000 GameState.DebugPrintPower() - SHOW_ENTITY - Updating Entity=[entityName=UNKNOWN ENTITY [cardType=INVALID] id=39 zone=DECK zonePos=0 cardId= player=2] CardID=EDR_528",
		"D 00:00:00.0000000 GameState.DebugPrintPower() -     tag=ZONE value=HAND",
	}

	var events []game.Event
	for _, line := range lines {
		parsedEvents, err := parser.ParseLine(line)
		if err != nil {
			t.Fatalf("parse line: %v", err)
		}
		events = append(events, parsedEvents...)
	}

	if len(events) != 1 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	drawEvent, ok := events[0].(*game.CardDrawnEvent)
	if !ok {
		t.Fatalf("unexpected event type: %T", events[0])
	}
	if drawEvent.Player != types.PlayerTypePlayer {
		t.Fatalf("expected draw to belong to local player, got %s", drawEvent.Player)
	}
}

func TestZoneParserResolvesLateCardRevealAfterDeckToHandTag(t *testing.T) {
	t.Parallel()

	parser := NewZoneParser()
	lines := []string{
		"D 00:00:00.0000000 GameState.DebugPrintGame() - PlayerID=1, PlayerName=UNKNOWN HUMAN PLAYER",
		"D 00:00:01.0000000 GameState.DebugPrintPower() -     TAG_CHANGE Entity=[entityName=UNKNOWN ENTITY [cardType=INVALID] id=22 zone=DECK zonePos=0 cardId= player=1] tag=ZONE value=HAND",
		"D 00:00:05.0000000 GameState.DebugPrintPower() -     SHOW_ENTITY - Updating Entity=[entityName=UNKNOWN ENTITY [cardType=INVALID] id=22 zone=HAND zonePos=3 cardId= player=1] CardID=END_033",
	}

	var events []game.Event
	for _, line := range lines {
		parsedEvents, err := parser.ParseLine(line)
		if err != nil {
			t.Fatalf("parse line: %v", err)
		}
		events = append(events, parsedEvents...)
	}

	if len(events) != 1 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	drawEvent, ok := events[0].(*game.CardDrawnEvent)
	if !ok {
		t.Fatalf("unexpected event type: %T", events[0])
	}
	if drawEvent.Player != types.PlayerTypePlayer {
		t.Fatalf("expected local draw, got %s", drawEvent.Player)
	}
	if drawEvent.Card == nil || drawEvent.Card.ID != "END_033" || drawEvent.Card.EntityID != 22 {
		t.Fatalf("unexpected drawn card: %#v", drawEvent.Card)
	}
}

func TestZoneParserDoesNotDoubleEmitSameDrawAcrossDuplicatePowerStreams(t *testing.T) {
	t.Parallel()

	parser := NewZoneParser()
	lines := []string{
		"D 00:00:00.0000000 GameState.DebugPrintPower() - SHOW_ENTITY - Updating Entity=[entityName=UNKNOWN ENTITY [cardType=INVALID] id=39 zone=DECK zonePos=0 cardId= player=2] CardID=EDR_528",
		"D 00:00:00.0000000 GameState.DebugPrintPower() -     tag=ZONE value=HAND",
		"D 00:00:00.0000000 PowerTaskList.DebugPrintPower() - SHOW_ENTITY - Updating Entity=[entityName=UNKNOWN ENTITY [cardType=INVALID] id=39 zone=DECK zonePos=0 cardId= player=2] CardID=EDR_528",
		"D 00:00:00.0000000 PowerTaskList.DebugPrintPower() -     tag=ZONE value=HAND",
	}

	var events []game.Event
	for _, line := range lines {
		parsedEvents, err := parser.ParseLine(line)
		if err != nil {
			t.Fatalf("parse line: %v", err)
		}
		events = append(events, parsedEvents...)
	}

	if len(events) != 1 {
		t.Fatalf("expected single draw event, got %d", len(events))
	}
}
