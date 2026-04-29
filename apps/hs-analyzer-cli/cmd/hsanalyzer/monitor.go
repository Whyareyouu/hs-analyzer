package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"hsanalyzer/apps/hs-analyzer-cli/internal/delivery"
	"hsanalyzer/apps/hs-analyzer-cli/internal/liveapi"
	"hsanalyzer/apps/hs-analyzer-cli/internal/matchpackage"
	"hsanalyzer/internal/analyzer"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/config"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/logreader"
	"hsanalyzer/internal/monitor"
	"hsanalyzer/internal/parser"
	"hsanalyzer/pkg/contracts/overlay"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	defaultMonitorStatusInterval = 5 * time.Second
	defaultLiveAPIAddr           = "127.0.0.1:8765"
)

type monitorOptions struct {
	configPath             string
	logsPath               string
	cardsDBPath            string
	stateFile              string
	liveCardStatsFile      string
	overlaySnapshotFile    string
	liveAPIAddr            string
	matchUploadURL         string
	matchSpoolDir          string
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
	cmd.Flags().StringVar(&options.overlaySnapshotFile, "overlay-snapshot-file", "", "Path to overlay snapshot file")
	cmd.Flags().StringVar(&options.liveAPIAddr, "live-api-addr", defaultLiveAPIAddr, "Address for local live overlay API server")
	cmd.Flags().StringVar(&options.matchUploadURL, "match-upload-url", "", "Backend URL for POST /api/v1/matches")
	cmd.Flags().StringVar(&options.matchSpoolDir, "match-spool-dir", "", "Directory for durable pending match uploads")
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
	overlaySnapshotPath, err := resolveOverlaySnapshotPath(options.overlaySnapshotFile)
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

	persister, err := newMonitorMatchPersister(options, logger)
	if err != nil {
		return err
	}
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

	overlayFileSink, err := monitor.NewAtomicFileOverlaySnapshotSink(overlaySnapshotPath)
	if err != nil {
		return fmt.Errorf("create overlay snapshot sink: %w", err)
	}
	liveServer := liveapi.NewServer(overlay.NewSnapshot())
	overlaySink := multiOverlaySnapshotSink{
		sinks: []monitor.OverlaySnapshotSink{
			overlayFileSink,
			liveServer,
		},
	}
	overlayProjector := monitor.NewOverlaySnapshotProjector(monitor.OverlaySnapshotProjectorOptions{
		CardDB:        cardDB,
		AssetResolver: overlayCardAssetResolver{},
		Sink:          overlaySink,
	})
	if runtimeState.Overlay != nil && canRestoreLiveCardStats(runtimeState, powerLogPath, loadingLogPath) {
		if err := overlayProjector.Restore(*runtimeState.Overlay); err != nil {
			return fmt.Errorf("restore overlay projector: %w", err)
		}
	}

	eventCh := make(chan game.Event, 256)
	checkpointCh := make(chan monitorReaderUpdate, 64)
	errCh := make(chan error, 2)
	var wg sync.WaitGroup

	liveListener, err := net.Listen("tcp", strings.TrimSpace(options.liveAPIAddr))
	if err != nil {
		return fmt.Errorf("listen local live api %q: %w", options.liveAPIAddr, err)
	}
	defer func() {
		_ = liveListener.Close()
	}()
	liveHTTPServer := &http.Server{Handler: liveServer.Routes()}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = liveHTTPServer.Shutdown(shutdownCtx)
	}()
	go func() {
		if err := liveHTTPServer.Serve(liveListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case errCh <- fmt.Errorf("live api server: %w", err):
			case <-runCtx.Done():
			}
		}
	}()
	fmt.Fprintf(cmd.OutOrStdout(), "live api started: http://%s\n", liveListener.Addr().String())

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
			if err := overlayProjector.Process(event); err != nil {
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
			if err := overlayProjector.SetHealth(buildOverlayHealth(snapshot)); err != nil {
				cancelRun()
				wg.Wait()
				return err
			}
			if err := retryPending(runCtx, cmd, postGame, health); err != nil {
				cancelRun()
				wg.Wait()
				return err
			}
			runtimeState.LiveCardStats = projector.PersistenceSnapshot()
			runtimeState.Overlay = overlayProjector.PersistenceSnapshot()
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
			runtimeState.Overlay = overlayProjector.PersistenceSnapshot()
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

func resolveOverlaySnapshotPath(path string) (string, error) {
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
	return filepath.Join(projectRoot, "output", "overlay_snapshot.json"), nil
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

func buildOverlayHealth(snapshot monitor.LiveHealthSnapshot) overlay.Health {
	health := overlay.Health{
		Monitor:   overlay.HealthStateHealthy,
		Transport: overlay.HealthStateHealthy,
	}
	if snapshot.Watcher != monitor.WatcherHealthHealthy || snapshot.Storage != monitor.StorageHealthHealthy {
		health.Monitor = overlay.HealthStateDegraded
	}
	return health
}

type overlayCardAssetResolver struct{}

func (overlayCardAssetResolver) ResolveCardImage(cardID string, _ int) string {
	cardID = strings.TrimSpace(cardID)
	if cardID == "" {
		return ""
	}
	return fmt.Sprintf("https://art.hearthstonejson.com/v1/render/latest/ruRU/256x/%s.png", cardID)
}

type multiOverlaySnapshotSink struct {
	sinks []monitor.OverlaySnapshotSink
}

func (s multiOverlaySnapshotSink) Publish(snapshot overlay.Snapshot) error {
	for _, sink := range s.sinks {
		if sink == nil {
			continue
		}
		if err := sink.Publish(snapshot); err != nil {
			return err
		}
	}
	return nil
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

type monitorMatchPersister interface {
	monitor.MatchPersister
	SavedCount() int
}

func newMonitorMatchPersister(options monitorOptions, logger *zap.Logger) (monitorMatchPersister, error) {
	if strings.TrimSpace(options.matchUploadURL) == "" {
		return &runtimeMatchPersister{logger: logger}, nil
	}

	spoolDir := strings.TrimSpace(options.matchSpoolDir)
	if spoolDir == "" {
		workingDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
		projectRoot, err := findProjectRoot(workingDir)
		if err != nil {
			return nil, err
		}
		spoolDir = filepath.Join(projectRoot, "output", "match_upload_spool")
	}

	spool, err := delivery.NewFileSpool(spoolDir, time.Now)
	if err != nil {
		return nil, fmt.Errorf("create match upload spool: %w", err)
	}

	return &deliveryMatchPersister{
		logger: logger,
		spool:  spool,
		uploader: delivery.HTTPUploader{
			Endpoint: options.matchUploadURL,
		},
	}, nil
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

type deliveryMatchPersister struct {
	logger   *zap.Logger
	spool    *delivery.FileSpool
	uploader delivery.Uploader

	mu        sync.Mutex
	delivered int
}

func (p *deliveryMatchPersister) SaveMatch(ctx context.Context, match monitor.PersistedMatch) error {
	if p == nil {
		return fmt.Errorf("delivery match persister is nil")
	}

	payload := matchpackage.BuildMatchUploadPayload(match, appVersion)
	item, err := p.spool.Enqueue(payload)
	if err != nil {
		return fmt.Errorf("enqueue match upload: %w", err)
	}

	results, err := p.spool.Deliver(ctx, p.uploader)
	if err != nil {
		return fmt.Errorf("deliver match uploads: %w", err)
	}
	for _, result := range results {
		if result.SpoolID != item.SpoolID {
			continue
		}
		if result.Status == delivery.StatusDelivered {
			p.mu.Lock()
			p.delivered++
			delivered := p.delivered
			p.mu.Unlock()
			p.logger.Info("uploaded live match",
				zap.Int("uploaded_total", delivered),
				zap.String("match_key", match.Key),
				zap.String("idempotency_key", item.IdempotencyKey),
			)
			return nil
		}
		return fmt.Errorf("match upload not delivered: status=%s err=%s", result.Status, result.LastError)
	}

	return fmt.Errorf("match upload result not found for %s", item.SpoolID)
}

func (p *deliveryMatchPersister) SavedCount() int {
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.delivered
}
