package analyzer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/config"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/logreader"
	"hsanalyzer/internal/parser"
	"hsanalyzer/internal/state"
)

// ReaderFactory создает reader для чтения логов.
type ReaderFactory func(bufferSize int) logreader.LogReader

// Analyzer оркестрирует чтение логов, парсинг событий и применение их к состоянию игры.
type Analyzer struct {
	config         *config.Config
	cardDB         carddb.CardDatabase
	parserRegistry *parser.Registry
	pipeline       *Pipeline
	readerFactory  ReaderFactory
}

// NewAnalyzer создает основной analyzer и валидирует его зависимости.
func NewAnalyzer(
	cfg *config.Config,
	db carddb.CardDatabase,
	registry *parser.Registry,
	pipeline *Pipeline,
	readerFactory ReaderFactory,
) (*Analyzer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if db == nil {
		return nil, fmt.Errorf("card database is nil")
	}
	if registry == nil {
		return nil, fmt.Errorf("parser registry is nil")
	}
	if pipeline == nil {
		return nil, fmt.Errorf("pipeline is nil")
	}
	if readerFactory == nil {
		readerFactory = func(bufferSize int) logreader.LogReader {
			return logreader.NewFileReader(bufferSize)
		}
	}

	return &Analyzer{
		config:         cfg,
		cardDB:         db,
		parserRegistry: registry,
		pipeline:       pipeline,
		readerFactory:  readerFactory,
	}, nil
}

// Analyze выполняет основной цикл анализа: загрузка БД карт, чтение логов, парсинг и применение событий.
func (a *Analyzer) Analyze(ctx context.Context) (*state.State, error) {
	if a == nil {
		return nil, fmt.Errorf("analyzer is nil")
	}

	if err := a.cardDB.Load(a.config.CardsDBPath); err != nil {
		return nil, fmt.Errorf("load card database: %w", err)
	}

	events, err := a.readAndParseLogs(ctx)
	if err != nil {
		return nil, err
	}

	sortEvents(events)

	gameState := state.New("", state.GameModeUnknown)
	if err := a.pipeline.ProcessAll(ctx, gameState, events); err != nil {
		return nil, fmt.Errorf("process events: %w", err)
	}

	return gameState, nil
}

func (a *Analyzer) readAndParseLogs(ctx context.Context) ([]game.Event, error) {
	parsers := a.parserRegistry.List()
	sort.SliceStable(parsers, func(i, j int) bool {
		return parsers[i].LogType() < parsers[j].LogType()
	})

	events := make([]game.Event, 0)
	for _, logParser := range parsers {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		logPath, ok, err := a.resolveLogPath(logParser.LogType())
		if err != nil {
			return nil, fmt.Errorf("resolve log path for %s: %w", logParser.LogType(), err)
		}
		if !ok {
			continue
		}

		content, err := a.readLogContent(logPath)
		if err != nil {
			return nil, fmt.Errorf("read %s log %q: %w", logParser.LogType(), logPath, err)
		}

		parsedEvents, err := logParser.Parse(ctx, strings.NewReader(content))
		if err != nil {
			return nil, fmt.Errorf("parse %s log %q: %w", logParser.LogType(), logPath, err)
		}

		events = append(events, parsedEvents...)
	}

	return events, nil
}

func (a *Analyzer) resolveLogPath(logType parser.LogType) (string, bool, error) {
	for _, candidate := range logFileCandidates(logType) {
		path := filepath.Join(a.config.LogsPath, candidate)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return path, true, nil
		}
		if err != nil && !os.IsNotExist(err) {
			return "", false, err
		}
	}

	return "", false, nil
}

func (a *Analyzer) readLogContent(path string) (string, error) {
	reader := a.readerFactory(a.config.Performance.BufferSize)
	if reader == nil {
		return "", fmt.Errorf("reader factory returned nil")
	}
	if err := reader.Open(path); err != nil {
		return "", err
	}
	defer func() {
		_ = reader.Close()
	}()

	var builder strings.Builder
	for {
		line, err := reader.Next()
		if err == nil {
			if builder.Len() > 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString(line)
			continue
		}

		if logreader.IsEOF(err) || err == io.EOF {
			break
		}

		return "", err
	}

	return builder.String(), nil
}

func sortEvents(events []game.Event) {
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Timestamp().Before(events[j].Timestamp())
	})
}

func logFileCandidates(logType parser.LogType) []string {
	switch logType {
	case parser.LogTypePower:
		return []string{"Power.log"}
	case parser.LogTypeZone:
		return []string{"Zone.log"}
	case parser.LogTypeLoadingScreen:
		return []string{"LoadingScreen.log", "LoadingScreen_old.log"}
	case parser.LogTypeDeck:
		return []string{"Decks.log"}
	default:
		return nil
	}
}
