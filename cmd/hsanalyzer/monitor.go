package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"hsanalyzer/internal/analyzer"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/config"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/logreader"
	"hsanalyzer/internal/monitor"
	"hsanalyzer/internal/parser"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	defaultMonitorStatusInterval = 5 * time.Second
)

type monitorOptions struct {
	configPath             string
	logsPath               string
	cardsDBPath            string
	stateFile              string
	liveCardStatsFile      string
	pollInterval           time.Duration
	statusInterval         time.Duration
	persistIncompleteFlush bool
	runFor                 time.Duration
}

type monitorReaderUpdate struct {
	source     string
	checkpoint monitorReaderCheckpoint
}

func newMonitorCmd() *cobra.Command {
	options := monitorOptions{}

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor Hearthstone logs in live mode",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMonitor(cmd.Context(), cmd, options)
		},
	}

	cmd.Flags().StringVar(&options.configPath, "config", "", "Path to config file")
	cmd.Flags().StringVar(&options.logsPath, "logs-path", "", "Path to logs directory or active Power.log file")
	cmd.Flags().StringVar(&options.cardsDBPath, "cards-db-path", "", "Path to cards database")
	cmd.Flags().StringVar(&options.stateFile, "state-file", "", "Path to monitor runtime state file for restart recovery")
	cmd.Flags().StringVar(&options.liveCardStatsFile, "live-card-stats-file", "", "Path to live card stats snapshot file")
	cmd.Flags().DurationVar(&options.pollInterval, "poll-interval", 250*time.Millisecond, "Live polling interval for growing log files")
	cmd.Flags().DurationVar(&options.statusInterval, "status-interval", defaultMonitorStatusInterval, "Interval for status output")
	cmd.Flags().BoolVar(&options.persistIncompleteFlush, "persist-incomplete-flush", false, "Persist incomplete session on shutdown flush")
	cmd.Flags().DurationVar(&options.runFor, "run-for", 0, "Stop monitoring after this duration (0 means run until Ctrl+C)")

	return cmd
}

func runMonitor(parentCtx context.Context, cmd *cobra.Command, options monitorOptions) error {
	logger := loggerFromContext(parentCtx)

	cfg, err := loadCLIConfig(options.configPath)
	if err != nil {
		return err
	}
	if err := applyMonitorOverrides(cfg, options); err != nil {
		return err
	}
	if err := validateMonitorConfig(cfg); err != nil {
		return err
	}

	powerLogPath, loadingLogPath := resolveLiveLogPaths(cfg)
	fmt.Fprintf(cmd.OutOrStdout(), "monitoring started: power=%s loading=%s\n", powerLogPath, loadingLogPath)

	statePath := strings.TrimSpace(options.stateFile)
	if statePath == "" {
		var err error
		statePath, err = defaultMonitorStatePath()
		if err != nil {
			return err
		}
	}
	runtimeState, err := loadMonitorRuntimeState(statePath)
	if err != nil {
		return err
	}
	liveCardStatsPath, err := resolveLiveCardStatsPath(options.liveCardStatsFile)
	if err != nil {
		return err
	}

	baseCtx := parentCtx
	if options.runFor > 0 {
		timeoutCtx, cancel := context.WithTimeout(baseCtx, options.runFor)
		defer cancel()
		baseCtx = timeoutCtx
	}

	signalCtx, stop := signal.NotifyContext(baseCtx, os.Interrupt)
	defer stop()

	runCtx, cancelRun := context.WithCancel(signalCtx)
	defer cancelRun()

	pipeline := analyzer.NewPipeline()
	if err := pipeline.RegisterHandler(analyzer.NewApplyEventHandler()); err != nil {
		return fmt.Errorf("register apply handler: %w", err)
	}

	cardDB := carddb.NewJSONDatabase()
	if err := cardDB.Load(cfg.CardsDBPath); err != nil {
		return fmt.Errorf("load cards database: %w", err)
	}

	builder, err := monitor.NewStateBuilder(pipeline)
	if err != nil {
		return fmt.Errorf("create state builder: %w", err)
	}

	persister := &runtimeMatchPersister{logger: logger}
	postGame, err := monitor.NewPostGamePipeline(builder, persister, monitor.PostGamePipelineOptions{
		PersistIncompleteFlush: options.persistIncompleteFlush,
	})
	if err != nil {
		return fmt.Errorf("create post-game pipeline: %w", err)
	}

	detector := monitor.NewSessionDetector()
	health := monitor.NewHealthTracker()
	powerParser := parser.NewPowerParser()
	loadingParser := parser.NewLoadingScreenParser()
	liveCardStatsSink, err := monitor.NewAtomicFileLiveCardStatsSink(liveCardStatsPath)
	if err != nil {
		return fmt.Errorf("create live card stats sink: %w", err)
	}
	projector := monitor.NewLiveCardStatsProjector(monitor.LiveCardStatsProjectorOptions{
		CardDB:        cardDB,
		AssetResolver: monitor.NoopCardAssetResolver{},
		Sink:          liveCardStatsSink,
	})
	if runtimeState.LiveCardStats != nil && canRestoreLiveCardStats(runtimeState, powerLogPath, loadingLogPath) {
		if err := projector.Restore(*runtimeState.LiveCardStats); err != nil {
			return fmt.Errorf("restore live card stats projector: %w", err)
		}
	}

	eventCh := make(chan game.Event, 256)
	checkpointCh := make(chan monitorReaderUpdate, 64)
	errCh := make(chan error, 2)
	var wg sync.WaitGroup

	startLiveReader := func(path string, source string, parseLine func(string) ([]game.Event, error), startOffset int64) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			reader := logreader.NewLiveFileReader(logreader.LiveReaderConfig{
				BufferSize:   cfg.Performance.BufferSize,
				PollInterval: options.pollInterval,
				StartOffset:  startOffset,
			})
			if err := reader.Open(path); err != nil {
				select {
				case errCh <- fmt.Errorf("%s reader open: %w", source, err):
				case <-runCtx.Done():
				}
				return
			}
			defer func() {
				_ = reader.Close()
			}()

			for {
				line, err := reader.Next(runCtx)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || runCtx.Err() != nil {
						return
					}
					health.SetWatcherError(err)
					select {
					case errCh <- fmt.Errorf("%s reader: %w", source, err):
					case <-runCtx.Done():
					}
					return
				}

				parsedEvents, err := parseLine(line)
				if err != nil {
					health.SetWatcherError(err)
					select {
					case errCh <- fmt.Errorf("%s parser: %w", source, err):
					case <-runCtx.Done():
					}
					return
				}
				health.SetWatcherHealthy()
				select {
				case checkpointCh <- monitorReaderUpdate{
					source: source,
					checkpoint: monitorReaderCheckpoint{
						Path:      reader.State().Path,
						Offset:    reader.State().Offset,
						UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
					},
				}:
				default:
				}

				for _, event := range parsedEvents {
					if event == nil {
						continue
					}
					select {
					case eventCh <- event:
					case <-runCtx.Done():
						return
					}
				}
			}
		}()
	}

	startLiveReader(powerLogPath, "power", powerParser.ParseLine, resolveReaderStartOffset(runtimeState, "power", powerLogPath))
	startLiveReader(loadingLogPath, "loading_screen", loadingParser.ParseLine, resolveReaderStartOffset(runtimeState, "loading_screen", loadingLogPath))

	statusInterval := options.statusInterval
	if statusInterval <= 0 {
		statusInterval = defaultMonitorStatusInterval
	}
	statusTicker := time.NewTicker(statusInterval)
	defer statusTicker.Stop()

	for {
		select {
		case event := <-eventCh:
			before := projector.Snapshot()
			if err := projector.Process(event); err != nil {
				cancelRun()
				wg.Wait()
				return err
			}
			if _, ok := event.(*game.CardPlayedEvent); ok {
				after := projector.Snapshot()
				if after.ProjectionVersion > before.ProjectionVersion {
					fmt.Fprintf(
						cmd.OutOrStdout(),
						"played cards: self=%s opponent=%s\n",
						formatLiveCardStatsCardsForTerminal(after.Players.Self.Cards),
						formatLiveCardStatsCardsForTerminal(after.Players.Opponent.Cards),
					)
				}
			}
			for _, session := range detector.Process(event) {
				if err := finalizeLiveSession(runCtx, cmd, postGame, health, session); err != nil {
					cancelRun()
					wg.Wait()
					return err
				}
			}

		case update := <-checkpointCh:
			if runtimeState.Readers == nil {
				runtimeState.Readers = make(map[string]monitorReaderCheckpoint)
			}
			runtimeState.Readers[update.source] = update.checkpoint

		case err := <-errCh:
			cancelRun()
			wg.Wait()
			return err

		case <-statusTicker.C:
			health.SetPathMissing(liveLogPathMissing(powerLogPath, loadingLogPath))
			state := detector.State()
			health.SetSessionState(state)
			selfCheck := health.SelfCheck()
			snapshot := health.Snapshot()
			fmt.Fprintf(
				cmd.OutOrStdout(),
				"status: phase=%s pending=%d active=%d persisted=%d watcher=%s storage=%s healthy=%t\n",
				state.Phase,
				state.PendingCount,
				state.ActiveEventCount,
				persister.SavedCount(),
				snapshot.Watcher,
				snapshot.Storage,
				selfCheck.Healthy,
			)
			logger.Info("monitor status",
				zap.String("phase", string(state.Phase)),
				zap.Int("pending_count", state.PendingCount),
				zap.Int("active_event_count", state.ActiveEventCount),
				zap.Int("persisted_count", persister.SavedCount()),
				zap.String("watcher_health", string(snapshot.Watcher)),
				zap.String("storage_health", string(snapshot.Storage)),
				zap.Bool("healthy", selfCheck.Healthy),
				zap.Strings("issues", selfCheck.Issues),
			)
			if err := retryPending(runCtx, cmd, postGame, health); err != nil {
				cancelRun()
				wg.Wait()
				return err
			}
			runtimeState.LiveCardStats = projector.PersistenceSnapshot()
			if err := saveMonitorRuntimeState(statePath, runtimeState); err != nil {
				cancelRun()
				wg.Wait()
				return err
			}

		case <-runCtx.Done():
			flushReason := monitor.CompletionReasonFlush
			if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
				flushReason = monitor.CompletionReasonStreamEnd
			}

			if session, ok := detector.Flush(flushReason); ok {
				if err := finalizeLiveSession(context.Background(), cmd, postGame, health, session); err != nil {
					cancelRun()
					wg.Wait()
					return err
				}
			}
			if err := retryPending(context.Background(), cmd, postGame, health); err != nil {
				cancelRun()
				wg.Wait()
				return err
			}
			runtimeState.LiveCardStats = projector.PersistenceSnapshot()
			if err := saveMonitorRuntimeState(statePath, runtimeState); err != nil {
				cancelRun()
				wg.Wait()
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "monitoring stopped: persisted=%d\n", persister.SavedCount())
			cancelRun()
			wg.Wait()
			return nil
		}
	}
}

func applyMonitorOverrides(cfg *config.Config, options monitorOptions) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// In live mode we want "no --logs-path" to mean auto-discovery, not the
	// repository-local default "logs" directory from DefaultConfig.
	if strings.TrimSpace(options.configPath) == "" && strings.TrimSpace(options.logsPath) == "" {
		cfg.LogsPath = ""
		cfg.PowerLogPath = ""
	}

	if strings.TrimSpace(options.logsPath) != "" {
		cfg.LogsPath = resolveCLIPath(workingDir, options.logsPath)
		cfg.PowerLogPath = ""
	}
	if strings.TrimSpace(options.cardsDBPath) != "" {
		cfg.CardsDBPath = resolveCLIPath(workingDir, options.cardsDBPath)
	}

	if err := config.NormalizeLogSource(cfg, config.LogSourceOptions{
		LogsPathExplicit:   strings.TrimSpace(options.logsPath) != "",
		AllowAutoDiscovery: true,
	}); err != nil {
		return fmt.Errorf("invalid logs path: %w", err)
	}

	return nil
}

func validateMonitorConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if err := validateDirectory(cfg.LogsPath); err != nil {
		return err
	}
	return validateFile(cfg.CardsDBPath)
}

func resolveLiveLogPaths(cfg *config.Config) (string, string) {
	powerLogPath := strings.TrimSpace(cfg.PowerLogPath)
	if powerLogPath == "" {
		powerLogPath = filepath.Join(cfg.LogsPath, "Power.log")
	}
	loadingLogPath := filepath.Join(cfg.LogsPath, "LoadingScreen.log")
	return powerLogPath, loadingLogPath
}

func resolveLiveCardStatsPath(path string) (string, error) {
	if strings.TrimSpace(path) != "" {
		workingDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}
		return resolveCLIPath(workingDir, path), nil
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	projectRoot, err := findProjectRoot(workingDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, "output", "live_card_stats.json"), nil
}

func canRestoreLiveCardStats(state monitorRuntimeState, powerLogPath, loadingLogPath string) bool {
	checks := map[string]string{
		"power":          powerLogPath,
		"loading_screen": loadingLogPath,
	}

	for source, currentPath := range checks {
		checkpoint, ok := state.Readers[source]
		if !ok || strings.TrimSpace(checkpoint.Path) == "" {
			continue
		}
		if filepath.Clean(checkpoint.Path) != filepath.Clean(currentPath) {
			return false
		}
	}

	return true
}

func formatLiveCardStatsCardsForTerminal(cards []monitor.LiveCardStatsCardEntry) string {
	if len(cards) == 0 {
		return "[]"
	}

	items := make([]string, 0, len(cards))
	for _, current := range cards {
		name := strings.TrimSpace(current.Card.Name)
		if name == "" {
			name = current.CardID
		}
		items = append(items, fmt.Sprintf("%s x%d", name, current.PlayCount))
	}

	return "[" + strings.Join(items, ", ") + "]"
}

func finalizeLiveSession(
	ctx context.Context,
	cmd *cobra.Command,
	postGame *monitor.PostGamePipeline,
	health *monitor.HealthTracker,
	session monitor.CompletedSession,
) error {
	result, err := postGame.FinalizeSession(ctx, session)
	if err != nil {
		health.SetStorageError(err)
		return fmt.Errorf("finalize session: %w", err)
	}

	switch {
	case result.Persisted:
		health.SetStorageHealthy()
		fmt.Fprintf(cmd.OutOrStdout(), "match persisted: key=%s complete=%t reason=%s\n", result.Match.Key, result.Match.Complete, result.Match.Reason)
	case result.Queued:
		health.SetStorageError(result.PersistError)
		fmt.Fprintf(cmd.OutOrStdout(), "match queued: key=%s err=%v\n", result.Match.Key, result.PersistError)
	case result.Duplicate:
		fmt.Fprintf(cmd.OutOrStdout(), "match duplicate: key=%s\n", result.Match.Key)
	case result.Skipped:
		fmt.Fprintf(cmd.OutOrStdout(), "match skipped: complete=%t reason=%s\n", result.Match.Complete, result.Match.Reason)
	}

	return nil
}

func retryPending(ctx context.Context, cmd *cobra.Command, postGame *monitor.PostGamePipeline, health *monitor.HealthTracker) error {
	results, err := postGame.RetryPending(ctx)
	if err != nil {
		health.SetStorageError(err)
		return fmt.Errorf("retry pending matches: %w", err)
	}

	for _, result := range results {
		if result.Persisted {
			health.SetStorageHealthy()
			fmt.Fprintf(cmd.OutOrStdout(), "match persisted after retry: key=%s\n", result.Match.Key)
			continue
		}
		if result.Queued && result.PersistError != nil {
			health.SetStorageError(result.PersistError)
			fmt.Fprintf(cmd.OutOrStdout(), "match still queued: key=%s err=%v\n", result.Match.Key, result.PersistError)
		}
	}

	return nil
}

func liveLogPathMissing(paths ...string) bool {
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return true
		}
	}
	return false
}

type runtimeMatchPersister struct {
	logger *zap.Logger

	mu    sync.Mutex
	saved int
}

func (p *runtimeMatchPersister) SaveMatch(_ context.Context, match monitor.PersistedMatch) error {
	if p == nil {
		return fmt.Errorf("runtime match persister is nil")
	}

	p.mu.Lock()
	p.saved++
	current := p.saved
	p.mu.Unlock()

	p.logger.Info("persisted live match",
		zap.Int("persisted_total", current),
		zap.String("match_key", match.Key),
		zap.Bool("complete", match.Complete),
		zap.String("reason", string(match.Reason)),
	)

	return nil
}

func (p *runtimeMatchPersister) SavedCount() int {
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.saved
}
