package parser

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
	"hsanalyzer/pkg/hslog"
)

// LoadingScreenParser извлекает начало игры и режим из LoadingScreen.log.
type LoadingScreenParser struct {
	currentMode state.GameMode
}

func NewLoadingScreenParser() *LoadingScreenParser {
	return &LoadingScreenParser{
		currentMode: state.GameModeUnknown,
	}
}

func (p *LoadingScreenParser) Name() string {
	return "loading_screen_parser"
}

func (p *LoadingScreenParser) LogType() LogType {
	return LogTypeLoadingScreen
}

func (p *LoadingScreenParser) Parse(ctx context.Context, reader io.Reader) ([]game.Event, error) {
	if reader == nil {
		return nil, fmt.Errorf("%s: reader is nil", p.Name())
	}

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	p.reset()
	events := make([]game.Event, 0)

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
		return nil, fmt.Errorf("%s: read loading log: %w", p.Name(), err)
	}

	return events, nil
}

// ParseLine parses one LoadingScreen.log line incrementally.
func (p *LoadingScreenParser) ParseLine(line string) ([]game.Event, error) {
	if p == nil {
		return nil, fmt.Errorf("loading screen parser is nil")
	}

	if matches := hslog.LoadingScreenBoardPattern.FindStringSubmatch(line); matches != nil {
		p.currentMode = mapBoardToGameMode(matches[1])
		return nil, nil
	}

	matches := hslog.LoadingScreenSceneLoadedPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, nil
	}

	prevMode := matches[1]
	nextMode := matches[2]
	if nextMode != hslog.SceneModeGameplay || prevMode == hslog.SceneModeGameplay {
		return nil, nil
	}

	event := game.NewGameStartEvent("", p.currentMode)
	p.currentMode = state.GameModeUnknown

	return []game.Event{event}, nil
}

func (p *LoadingScreenParser) reset() {
	if p == nil {
		return
	}
	p.currentMode = state.GameModeUnknown
}

func mapBoardToGameMode(board string) state.GameMode {
	switch board {
	case hslog.BoardStandardGame:
		return state.GameModeStandard
	case hslog.BoardBattlegroundsGame:
		return state.GameModeBattlegrounds
	default:
		return state.GameModeUnknown
	}
}
