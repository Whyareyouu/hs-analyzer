package parser

import (
	"context"
	"strings"
	"testing"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/types"
)

func TestPowerParserMetadata(t *testing.T) {
	t.Parallel()

	parser := NewPowerParser()
	if parser.Name() != "power_parser" {
		t.Fatalf("unexpected parser name: %s", parser.Name())
	}
	if parser.LogType() != LogTypePower {
		t.Fatalf("unexpected log type: %s", parser.LogType())
	}
}

func TestPowerParserParseBaseStructure(t *testing.T) {
	t.Parallel()

	gameTypeLine := findLogLine(t, "Power.log", "GameType=GT_RANKED")
	attackLine := findLogLine(t, "Power.log", "BLOCK_START BlockType=ATTACK Entity=[entityName=Хранитель стаи id=29")

	input := strings.Join([]string{
		gameTypeLine,
		attackLine,
	}, "\n")

	parser := NewPowerParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse power log: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("unexpected events count: %d", len(events))
	}
}

func TestPowerParserRecognizesRealPowerLogFragments(t *testing.T) {
	t.Parallel()

	parser := NewPowerParser()
	gameTypeLine := findLogLine(t, "Power.log", "GameType=GT_RANKED")
	attackLine := findLogLine(t, "Power.log", "BLOCK_START BlockType=ATTACK Entity=[entityName=Хранитель стаи id=29")
	turnLine := findLogLine(t, "Power.log", "TAG_CHANGE Entity=GameEntity tag=TURN value=6")

	if !parser.isRelevantLine(gameTypeLine) {
		t.Fatalf("expected game type line to be relevant")
	}
	if !parser.isRelevantLine(attackLine) {
		t.Fatalf("expected attack line to be relevant")
	}
	if !parser.isRelevantLine(turnLine) {
		t.Fatalf("expected turn line to be relevant")
	}
}

func TestPowerParserParseTurnTransitionsFromRealLogFragments(t *testing.T) {
	t.Parallel()

	turnLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=GameEntity tag=TURN value=6")
	opponentTurnLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=HappY#23785 tag=TURN value=3")
	turnDupLine := findLogLine(t, "Power.log", "PowerTaskList.DebugPrintPower() -     TAG_CHANGE Entity=UNKNOWN HUMAN PLAYER tag=TURN value=4")
	nextTurnGameEntityLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=GameEntity tag=TURN value=7")
	playerTurnLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=UNKNOWN HUMAN PLAYER tag=TURN value=4")

	input := strings.Join([]string{
		turnLine,
		opponentTurnLine,
		turnDupLine,
		nextTurnGameEntityLine,
		playerTurnLine,
	}, "\n")

	parser := NewPowerParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse power log: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	startOpponent, ok := events[0].(*game.TurnStartEvent)
	if !ok {
		t.Fatalf("unexpected first event type: %T", events[0])
	}
	if startOpponent.Player != types.PlayerTypeOpponent {
		t.Fatalf("unexpected first turn player: %s", startOpponent.Player)
	}

	if _, ok := events[1].(*game.TurnEndEvent); !ok {
		t.Fatalf("unexpected second event type: %T", events[1])
	}

	startPlayer, ok := events[2].(*game.TurnStartEvent)
	if !ok {
		t.Fatalf("unexpected third event type: %T", events[2])
	}
	if startPlayer.Player != types.PlayerTypePlayer {
		t.Fatalf("unexpected second turn player: %s", startPlayer.Player)
	}
}

func TestPowerParserParseCardPlayedFromRealLogFragments(t *testing.T) {
	t.Parallel()

	playStartLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() - BLOCK_START BlockType=PLAY Entity=[entityName=Карта сектантов id=60")
	playResolveLine := findLogLine(t, "Power.log", "PowerTaskList.DebugPrintPower() - BLOCK_START BlockType=POWER Entity=[entityName=Карта сектантов id=60 zone=PLAY")

	input := strings.Join([]string{
		playStartLine,
		playResolveLine,
	}, "\n")

	parser := NewPowerParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse power log: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	cardPlayed, ok := events[0].(*game.CardPlayedEvent)
	if !ok {
		t.Fatalf("unexpected event type: %T", events[0])
	}
	if cardPlayed.Player != types.PlayerTypeOpponent {
		t.Fatalf("unexpected player: %s", cardPlayed.Player)
	}
	if cardPlayed.Card == nil {
		t.Fatalf("expected played card")
	}
	if cardPlayed.Card.ID != "TLC_515" {
		t.Fatalf("unexpected card id: %s", cardPlayed.Card.ID)
	}
	if cardPlayed.Card.EntityID != 60 {
		t.Fatalf("unexpected entity id: %d", cardPlayed.Card.EntityID)
	}
	if cardPlayed.Position != 0 {
		t.Fatalf("unexpected position: %d", cardPlayed.Position)
	}
}
