package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/formatter"
	"hsanalyzer/internal/state"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type analyzeOptions struct {
	configPath  string
	logsPath    string
	cardsDBPath string
	format      string
	output      string
	outputFile  string
	language    string
	verbose     bool
}

func newAnalyzeCmd() *cobra.Command {
	options := analyzeOptions{}

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze Hearthstone logs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAnalyze(cmd.Context(), cmd, options)
		},
	}

	cmd.Flags().StringVar(&options.configPath, "config", "", "Path to config file")
	cmd.Flags().StringVar(&options.logsPath, "logs-path", "", "Path to logs directory")
	cmd.Flags().StringVar(&options.cardsDBPath, "cards-db-path", "", "Path to cards database")
	cmd.Flags().StringVar(&options.format, "format", "", "Output format")
	cmd.Flags().StringVar(&options.output, "output", "", "Output destination: stdout or file")
	cmd.Flags().StringVar(&options.outputFile, "output-file", "", "Output file path when destination=file")
	cmd.Flags().StringVar(&options.language, "lang", "", "Output language: ru or en")
	cmd.Flags().BoolVar(&options.verbose, "verbose", false, "Enable verbose output")

	return cmd
}

func runAnalyze(ctx context.Context, cmd *cobra.Command, options analyzeOptions) error {
	logger := loggerFromContext(ctx)
	logger.Info("starting analysis",
		zap.String("logs_path", options.logsPath),
		zap.String("cards_db_path", options.cardsDBPath),
		zap.String("config_path", options.configPath),
	)

	container, err := buildContainer(options)
	if err != nil {
		logger.Error("failed to build container", zap.Error(err))
		return err
	}

	gameState, err := container.analyzer.Analyze(ctx)
	if err != nil {
		logger.Error("analysis failed", zap.Error(err))
		return fmt.Errorf("analyze logs: %w", err)
	}

	output, err := container.formatter.Format(gameState)
	if err != nil {
		logger.Error("formatting failed", zap.Error(err))
		return fmt.Errorf("format result: %w", err)
	}

	if err := writeOutput(cmd, container.config.Output.Destination, container.config.Output.FilePath, output); err != nil {
		logger.Error("failed to write output", zap.Error(err))
		return err
	}

	timelinePath, err := saveTimelineJSON(container, gameState)
	if err != nil {
		logger.Error("failed to save timeline json", zap.Error(err))
		return fmt.Errorf("save timeline json: %w", err)
	}

	logger.Info("analysis completed",
		zap.Int("turns", gameState.GetTurnCount()),
		zap.Int("events", gameState.GetEventCount()),
		zap.String("output_destination", container.config.Output.Destination),
		zap.String("output_format", container.config.Output.Format),
		zap.String("timeline_json", timelinePath),
	)

	return nil
}

func writeOutput(cmd *cobra.Command, destination, filePath, content string) error {
	switch destination {
	case "stdout", "":
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), content); err != nil {
			return fmt.Errorf("write stdout: %w", err)
		}
		return nil
	case "file":
		if filePath == "" {
			return fmt.Errorf("output file path is empty")
		}
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output destination %q", destination)
	}
}

func saveTimelineJSON(cfg interface {
	GetLogsPath() string
	GetOutputDestination() string
	GetOutputFilePath() string
	GetCardDatabase() carddb.CardDatabase
}, gameState *state.State) (string, error) {
	if gameState == nil {
		return "", fmt.Errorf("game state is nil")
	}

	outputDir, err := resolveTimelineOutputDir(cfg)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create timeline directory: %w", err)
	}

	formatter := formatter.NewJSONFormatter(cfg.GetCardDatabase())
	content, err := formatter.Format(gameState)
	if err != nil {
		return "", fmt.Errorf("format timeline json: %w", err)
	}

	fileName := fmt.Sprintf("timeline_%s.json", time.Now().Format("20060102150405"))
	path := filepath.Join(outputDir, fileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write timeline json: %w", err)
	}

	return path, nil
}

func resolveTimelineOutputDir(cfg interface {
	GetLogsPath() string
	GetOutputDestination() string
	GetOutputFilePath() string
}) (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	return filepath.Join(workingDir, "output"), nil
}
