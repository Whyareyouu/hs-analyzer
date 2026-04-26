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
	"hsanalyzer/internal/types"
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

// SessionBounds описывает границы игровой сессии внутри набора событий.
type SessionBounds struct {
	StartIndex int
	EndIndex   int
	StartTime  string
	EndTime    string
	HasStart   bool
	HasEnd     bool
}

// MatchResult содержит восстановленное состояние одного матча и его границы в исходном потоке событий.
type MatchResult struct {
	Index  int
	State  *state.State
	Bounds SessionBounds
	Events []game.Event
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

	events, bounds := a.MergeAndSortEvents(events)
	matches, err := a.ApplyMatches(ctx, events)
	if err != nil {
		return nil, err
	}

	gameState := buildAggregateState(matches)
	if gameState == nil {
		gameState, err = a.ApplyEvents(ctx, events)
		if err != nil {
			return nil, err
		}
	}
	gameState.SetMetadata("session_bounds", bounds)
	gameState.SetMetadata("matches", collectMatchStates(matches))
	gameState.SetMetadata("match_results", matches)
	gameState.SetMetadata("match_count", len(matches))

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
	if logType == parser.LogTypePower && strings.TrimSpace(a.config.PowerLogPath) != "" {
		info, err := os.Stat(a.config.PowerLogPath)
		if err == nil && !info.IsDir() {
			return a.config.PowerLogPath, true, nil
		}
		if err != nil && !os.IsNotExist(err) {
			return "", false, err
		}
		return "", false, nil
	}

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
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		builder.Grow(int(info.Size()) + 1)
	}
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

// MergeAndSortEvents объединяет события, сортирует их по timestamp и вычисляет границы игровой сессии.
func (a *Analyzer) MergeAndSortEvents(eventGroups ...[]game.Event) ([]game.Event, SessionBounds) {
	merged := mergeEvents(eventGroups...)
	sortEvents(merged)
	return merged, determineSessionBounds(merged)
}

// ApplyEvents создает начальное состояние и применяет к нему события через pipeline.
func (a *Analyzer) ApplyEvents(ctx context.Context, events []game.Event) (*state.State, error) {
	if a == nil {
		return nil, fmt.Errorf("analyzer is nil")
	}

	gameState := state.New("", state.GameModeUnknown)
	if err := a.pipeline.ProcessAll(ctx, gameState, events); err != nil {
		return nil, fmt.Errorf("process events: %w", err)
	}

	return gameState, nil
}

// ApplyMatches сегментирует общий поток событий на отдельные матчи и строит состояние для каждого.
func (a *Analyzer) ApplyMatches(ctx context.Context, events []game.Event) ([]MatchResult, error) {
	if a == nil {
		return nil, fmt.Errorf("analyzer is nil")
	}

	sessions := splitEventsIntoMatches(events)
	if len(sessions) == 0 {
		return nil, nil
	}

	results := make([]MatchResult, 0, len(sessions))
	for index, sessionEvents := range sessions {
		gameState, err := a.ApplyEvents(ctx, sessionEvents)
		if err != nil {
			return nil, fmt.Errorf("process match %d: %w", index, err)
		}

		bounds := determineSessionBounds(sessionEvents)
		gameState.SetMetadata("session_bounds", bounds)
		gameState.SetMetadata("match_index", index)
		results = append(results, MatchResult{
			Index:  index,
			State:  gameState,
			Bounds: bounds,
			Events: append([]game.Event(nil), sessionEvents...),
		})
	}

	return results, nil
}

func mergeEvents(eventGroups ...[]game.Event) []game.Event {
	total := 0
	for _, group := range eventGroups {
		total += len(group)
	}

	merged := make([]game.Event, 0, total)
	for _, group := range eventGroups {
		merged = append(merged, group...)
	}

	return merged
}

func sortEvents(events []game.Event) {
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Timestamp().Before(events[j].Timestamp())
	})
}

func determineSessionBounds(events []game.Event) SessionBounds {
	bounds := SessionBounds{
		StartIndex: -1,
		EndIndex:   -1,
	}

	for i, event := range events {
		if !bounds.HasStart && event.Type() == types.EventTypeGameStart {
			bounds.StartIndex = i
			bounds.StartTime = event.Timestamp().Format("2006-01-02T15:04:05.000000000Z07:00")
			bounds.HasStart = true
		}
		if event.Type() == types.EventTypeGameEnd {
			bounds.EndIndex = i
			bounds.EndTime = event.Timestamp().Format("2006-01-02T15:04:05.000000000Z07:00")
			bounds.HasEnd = true
		}
	}

	if !bounds.HasStart && len(events) > 0 {
		bounds.StartIndex = 0
		bounds.StartTime = events[0].Timestamp().Format("2006-01-02T15:04:05.000000000Z07:00")
	}
	if !bounds.HasEnd && len(events) > 0 {
		bounds.EndIndex = len(events) - 1
		bounds.EndTime = events[len(events)-1].Timestamp().Format("2006-01-02T15:04:05.000000000Z07:00")
	}

	return bounds
}

func splitEventsIntoMatches(events []game.Event) [][]game.Event {
	if len(events) == 0 {
		return nil
	}

	sessions := make([][]game.Event, 0)
	pending := make([]game.Event, 0)
	current := make([]game.Event, 0)

	flushCurrent := func() {
		if len(current) == 0 {
			return
		}
		sessions = append(sessions, append([]game.Event(nil), current...))
		current = current[:0]
	}

	for _, event := range events {
		if event == nil {
			continue
		}

		switch event.Type() {
		case types.EventTypeGameStart:
			if len(current) > 0 {
				flushCurrent()
			}
			current = append(current[:0], pending...)
			current = append(current, event)
			pending = pending[:0]
		case types.EventTypeGameEnd:
			if len(current) == 0 {
				current = append(current, pending...)
				pending = pending[:0]
			}
			current = append(current, event)
			flushCurrent()
		default:
			if len(current) > 0 {
				current = append(current, event)
			} else {
				pending = append(pending, event)
			}
		}
	}

	if len(current) > 0 {
		flushCurrent()
	} else if len(pending) > 0 {
		sessions = append(sessions, append([]game.Event(nil), pending...))
	}

	return sessions
}

func collectMatchStates(results []MatchResult) []*state.State {
	if len(results) == 0 {
		return nil
	}

	matches := make([]*state.State, 0, len(results))
	for _, result := range results {
		if result.State != nil {
			matches = append(matches, result.State)
		}
	}
	return matches
}

func buildAggregateState(results []MatchResult) *state.State {
	if len(results) == 0 {
		return nil
	}

	last := results[len(results)-1].State
	if last == nil {
		return nil
	}
	return last
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
