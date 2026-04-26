package parser

import (
	"context"
	"strings"
	"testing"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
)

func TestLoadingScreenParserParseGameStart(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		findLogLine(t, "LoadingScreen.log", "LoadingScreen.DisableTransitionUnfriendlyStuff() - BoardStandardGame(Clone)"),
		findLogLine(t, "LoadingScreen.log", "LoadingScreen.OnSceneLoaded() - prevMode=LOGIN currMode=GAMEPLAY"),
	}, "\n")

	parser := NewLoadingScreenParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse loading log: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	event, ok := events[0].(*game.GameStartEvent)
	if !ok {
		t.Fatalf("unexpected event type: %T", events[0])
	}
	if event.GameMode != state.GameModeStandard {
		t.Fatalf("unexpected game mode: %s", event.GameMode)
	}
}

func TestLoadingScreenParserParseIgnoresNonGameplayTransitions(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		findLogLine(t, "LoadingScreen.log", "LoadingScreen.OnSceneLoaded() - prevMode=STARTUP currMode=LOGIN"),
		findLogLine(t, "LoadingScreen.log", "LoadingScreen.OnSceneLoaded() - prevMode=GAMEPLAY currMode=TOURNAMENT"),
	}, "\n")

	parser := NewLoadingScreenParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse loading log: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("unexpected events count: %d", len(events))
	}
}

func TestLoadingScreenParserParseLineSupportsIncrementalStream(t *testing.T) {
	t.Parallel()

	boardLine := findLogLine(t, "LoadingScreen.log", "LoadingScreen.DisableTransitionUnfriendlyStuff() - BoardStandardGame(Clone)")
	sceneLine := findLogLine(t, "LoadingScreen.log", "LoadingScreen.OnSceneLoaded() - prevMode=LOGIN currMode=GAMEPLAY")

	parser := NewLoadingScreenParser()

	events, err := parser.ParseLine(boardLine)
	if err != nil {
		t.Fatalf("parse board line: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("unexpected events count after board line: %d", len(events))
	}

	events, err = parser.ParseLine(sceneLine)
	if err != nil {
		t.Fatalf("parse scene line: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("unexpected events count after scene line: %d", len(events))
	}

	event, ok := events[0].(*game.GameStartEvent)
	if !ok {
		t.Fatalf("unexpected event type: %T", events[0])
	}
	if event.GameMode != state.GameModeStandard {
		t.Fatalf("unexpected game mode: %s", event.GameMode)
	}
}
