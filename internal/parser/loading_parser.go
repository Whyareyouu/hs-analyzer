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
type LoadingScreenParser struct{}

func NewLoadingScreenParser() *LoadingScreenParser {
	return &LoadingScreenParser{}
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

	events := make([]game.Event, 0)
	currentMode := state.GameModeUnknown

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := scanner.Text()

		if matches := hslog.LoadingScreenBoardPattern.FindStringSubmatch(line); matches != nil {
			currentMode = mapBoardToGameMode(matches[1])
			continue
		}

		matches := hslog.LoadingScreenSceneLoadedPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		prevMode := matches[1]
		nextMode := matches[2]
		if nextMode != hslog.SceneModeGameplay || prevMode == hslog.SceneModeGameplay {
			continue
		}

		events = append(events, game.NewGameStartEvent("", currentMode))
		currentMode = state.GameModeUnknown
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: read loading log: %w", p.Name(), err)
	}

	return events, nil
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
