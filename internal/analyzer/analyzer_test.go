package analyzer

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/config"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/logreader"
	"hsanalyzer/internal/parser"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

type stubCardDatabase struct {
	loadPath string
	loadErr  error
}

func (db *stubCardDatabase) Load(path string) error {
	db.loadPath = path
	return db.loadErr
}

func (db *stubCardDatabase) GetCard(string) (*carddb.CardInfo, bool) {
	return nil, false
}

func (db *stubCardDatabase) GetCardByDBFID(int) (*carddb.CardInfo, bool) {
	return nil, false
}

func (db *stubCardDatabase) GetCardByName(string) (*carddb.CardInfo, bool) {
	return nil, false
}

type stubParser struct {
	name    string
	logType parser.LogType
	parseFn func(context.Context, io.Reader) ([]game.Event, error)
}

func (p stubParser) Parse(ctx context.Context, reader io.Reader) ([]game.Event, error) {
	if p.parseFn != nil {
		return p.parseFn(ctx, reader)
	}
	return nil, nil
}

func (p stubParser) Name() string {
	return p.name
}

func (p stubParser) LogType() parser.LogType {
	return p.logType
}

type stubLogReader struct {
	openFn  func(string) error
	nextFn  func() (string, error)
	closeFn func() error
}

func (r *stubLogReader) Open(path string) error {
	if r.openFn != nil {
		return r.openFn(path)
	}
	return nil
}

func (r *stubLogReader) Next() (string, error) {
	if r.nextFn != nil {
		return r.nextFn()
	}
	return "", io.EOF
}

func (r *stubLogReader) Close() error {
	if r.closeFn != nil {
		return r.closeFn()
	}
	return nil
}

type testEvent struct {
	eventType types.EventType
	timestamp time.Time
	applyFn   func(*state.State) error
}

func (e testEvent) Type() types.EventType {
	return e.eventType
}

func (e testEvent) Timestamp() time.Time {
	return e.timestamp
}

func (e testEvent) Apply(gameState *state.State) error {
	if e.applyFn != nil {
		return e.applyFn(gameState)
	}
	return nil
}

func TestNewAnalyzerRejectsNilDependencies(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	db := &stubCardDatabase{}
	registry := parser.NewRegistry()
	pipeline := NewPipeline()

	cases := []struct {
		name         string
		cfg          *config.Config
		db           carddb.CardDatabase
		registry     *parser.Registry
		pipeline     *Pipeline
		expectErrSub string
	}{
		{name: "nil config", db: db, registry: registry, pipeline: pipeline, expectErrSub: "config is nil"},
		{name: "nil db", cfg: cfg, registry: registry, pipeline: pipeline, expectErrSub: "card database is nil"},
		{name: "nil registry", cfg: cfg, db: db, pipeline: pipeline, expectErrSub: "parser registry is nil"},
		{name: "nil pipeline", cfg: cfg, db: db, registry: registry, expectErrSub: "pipeline is nil"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewAnalyzer(tc.cfg, tc.db, tc.registry, tc.pipeline, nil)
			if err == nil || !strings.Contains(err.Error(), tc.expectErrSub) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAnalyzerAnalyzeLoadsCardDBReadsLogsAndProcessesEvents(t *testing.T) {
	t.Parallel()

	logsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(logsDir, "Power.log"), []byte("power line"), 0o644); err != nil {
		t.Fatalf("write Power.log: %v", err)
	}

	cfg := &config.Config{
		LogsPath:    logsDir,
		CardsDBPath: filepath.Join(logsDir, "cards.json"),
		Performance: config.PerformanceConfig{BufferSize: 128},
	}

	db := &stubCardDatabase{}
	registry := parser.NewRegistry()
	if err := registry.Register(stubParser{
		name:    "power_stub",
		logType: parser.LogTypePower,
		parseFn: func(_ context.Context, reader io.Reader) ([]game.Event, error) {
			data, err := io.ReadAll(reader)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(string(data)) != "power line" {
				t.Fatalf("unexpected parser input: %q", string(data))
			}

			return []game.Event{
				game.NewGameStartEvent("game-1", state.GameModeStandard),
				game.NewTurnStartEvent(types.PlayerTypePlayer),
			}, nil
		},
	}); err != nil {
		t.Fatalf("register parser: %v", err)
	}

	pipeline := NewPipeline()
	if err := pipeline.RegisterHandler(NewApplyEventHandler()); err != nil {
		t.Fatalf("register apply handler: %v", err)
	}

	analyzer, err := NewAnalyzer(cfg, db, registry, pipeline, nil)
	if err != nil {
		t.Fatalf("new analyzer: %v", err)
	}

	gameState, err := analyzer.Analyze(context.Background())
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}

	if db.loadPath != cfg.CardsDBPath {
		t.Fatalf("unexpected cards db path: %s", db.loadPath)
	}
	if gameState.GameID != "game-1" {
		t.Fatalf("unexpected game id: %s", gameState.GameID)
	}
	if gameState.GameMode != state.GameModeStandard {
		t.Fatalf("unexpected game mode: %s", gameState.GameMode)
	}
	if gameState.GetTurnCount() != 1 {
		t.Fatalf("unexpected turn count: %d", gameState.GetTurnCount())
	}
	if gameState.GetEventCount() != 2 {
		t.Fatalf("unexpected event count: %d", gameState.GetEventCount())
	}
}

func TestAnalyzerReadAndParseLogsUsesLoadingFallbackAndSkipsMissingLogs(t *testing.T) {
	t.Parallel()

	logsDir := t.TempDir()
	loadingPath := filepath.Join(logsDir, "LoadingScreen_old.log")
	if err := os.WriteFile(loadingPath, []byte("loading line"), 0o644); err != nil {
		t.Fatalf("write loading log: %v", err)
	}

	cfg := &config.Config{
		LogsPath:    logsDir,
		CardsDBPath: filepath.Join(logsDir, "cards.json"),
		Performance: config.PerformanceConfig{BufferSize: 128},
	}

	registry := parser.NewRegistry()
	if err := registry.Register(stubParser{
		name:    "loading_stub",
		logType: parser.LogTypeLoadingScreen,
		parseFn: func(_ context.Context, reader io.Reader) ([]game.Event, error) {
			data, err := io.ReadAll(reader)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(string(data)) != "loading line" {
				t.Fatalf("unexpected loading parser input: %q", string(data))
			}
			return []game.Event{game.NewGameStartEvent("fallback", state.GameModeStandard)}, nil
		},
	}); err != nil {
		t.Fatalf("register parser: %v", err)
	}

	if err := registry.Register(stubParser{
		name:    "zone_stub",
		logType: parser.LogTypeZone,
		parseFn: func(_ context.Context, _ io.Reader) ([]game.Event, error) {
			t.Fatalf("zone parser should be skipped when Zone.log is absent")
			return nil, nil
		},
	}); err != nil {
		t.Fatalf("register parser: %v", err)
	}

	analyzer, err := NewAnalyzer(cfg, &stubCardDatabase{}, registry, NewPipeline(), nil)
	if err != nil {
		t.Fatalf("new analyzer: %v", err)
	}

	events, err := analyzer.readAndParseLogs(context.Background())
	if err != nil {
		t.Fatalf("read and parse logs: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("unexpected events count: %d", len(events))
	}
}

func TestAnalyzerAnalyzeSortsEventsBeforePipeline(t *testing.T) {
	t.Parallel()

	logsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(logsDir, "Power.log"), []byte("power line"), 0o644); err != nil {
		t.Fatalf("write Power.log: %v", err)
	}

	cfg := &config.Config{
		LogsPath:    logsDir,
		CardsDBPath: filepath.Join(logsDir, "cards.json"),
		Performance: config.PerformanceConfig{BufferSize: 128},
	}

	oldest := time.Unix(10, 0)
	newest := time.Unix(20, 0)

	registry := parser.NewRegistry()
	if err := registry.Register(stubParser{
		name:    "power_stub",
		logType: parser.LogTypePower,
		parseFn: func(_ context.Context, _ io.Reader) ([]game.Event, error) {
			return []game.Event{
				testEvent{
					eventType: types.EventTypeTurnEnd,
					timestamp: newest,
					applyFn: func(gameState *state.State) error {
						order := gameState.GetMetadata("order")
						if order == nil {
							gameState.SetMetadata("order", []string{"new"})
							return nil
						}
						gameState.SetMetadata("order", append(order.([]string), "new"))
						return nil
					},
				},
				testEvent{
					eventType: types.EventTypeTurnStart,
					timestamp: oldest,
					applyFn: func(gameState *state.State) error {
						order := gameState.GetMetadata("order")
						if order == nil {
							gameState.SetMetadata("order", []string{"old"})
							return nil
						}
						gameState.SetMetadata("order", append(order.([]string), "old"))
						return nil
					},
				},
			}, nil
		},
	}); err != nil {
		t.Fatalf("register parser: %v", err)
	}

	pipeline := NewPipeline()
	if err := pipeline.RegisterHandler(NewApplyEventHandler()); err != nil {
		t.Fatalf("register apply handler: %v", err)
	}

	analyzer, err := NewAnalyzer(cfg, &stubCardDatabase{}, registry, pipeline, nil)
	if err != nil {
		t.Fatalf("new analyzer: %v", err)
	}

	gameState, err := analyzer.Analyze(context.Background())
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}

	order, _ := gameState.GetMetadata("order").([]string)
	if !reflect.DeepEqual(order, []string{"old", "new"}) {
		t.Fatalf("unexpected apply order: %v", order)
	}
}

func TestAnalyzerReadLogContentPropagatesReaderErrors(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		LogsPath:    t.TempDir(),
		CardsDBPath: filepath.Join(t.TempDir(), "cards.json"),
		Performance: config.PerformanceConfig{BufferSize: 128},
	}

	analyzer, err := NewAnalyzer(cfg, &stubCardDatabase{}, parser.NewRegistry(), NewPipeline(), func(int) logreader.LogReader {
		return &stubLogReader{
			openFn: func(string) error { return nil },
			nextFn: func() (string, error) { return "", errors.New("read failed") },
		}
	})
	if err != nil {
		t.Fatalf("new analyzer: %v", err)
	}

	_, err = analyzer.readLogContent("broken.log")
	if err == nil || !strings.Contains(err.Error(), "read failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
