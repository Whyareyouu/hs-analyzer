package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hsanalyzer/internal/analyzer"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/config"
	"hsanalyzer/internal/formatter"
	"hsanalyzer/internal/parser"
)

type appContainer struct {
	config    *config.Config
	cardDB    carddb.CardDatabase
	analyzer  *analyzer.Analyzer
	formatter formatter.Formatter
}

func (c *appContainer) GetLogsPath() string {
	if c == nil || c.config == nil {
		return ""
	}
	return c.config.LogsPath
}

func (c *appContainer) GetOutputDestination() string {
	if c == nil || c.config == nil {
		return ""
	}
	return c.config.Output.Destination
}

func (c *appContainer) GetOutputFilePath() string {
	if c == nil || c.config == nil {
		return ""
	}
	return c.config.Output.FilePath
}

func (c *appContainer) GetCardDatabase() carddb.CardDatabase {
	if c == nil {
		return nil
	}
	return c.cardDB
}

func buildContainer(options analyzeOptions) (*appContainer, error) {
	cfg, err := loadCLIConfig(options.configPath)
	if err != nil {
		return nil, err
	}

	if err := applyAnalyzeOverrides(cfg, options); err != nil {
		return nil, err
	}
	if err := validateCLIConfig(cfg); err != nil {
		return nil, err
	}

	db := carddb.NewJSONDatabase()

	parserRegistry := parser.NewRegistry()
	parserRegistry.MustRegister(parser.NewLoadingScreenParser())
	parserRegistry.MustRegister(parser.NewPowerParser())
	parserRegistry.MustRegister(parser.NewZoneParser())
	parserRegistry.MustRegister(parser.NewDeckParser(db))

	pipeline := analyzer.NewPipeline()
	if err := pipeline.RegisterHandler(analyzer.NewApplyEventHandler()); err != nil {
		return nil, fmt.Errorf("register apply handler: %w", err)
	}

	logAnalyzer, err := analyzer.NewAnalyzer(cfg, db, parserRegistry, pipeline, nil)
	if err != nil {
		return nil, err
	}

	formatterRegistry := formatter.NewRegistry()
	formatterRegistry.MustRegister(formatter.NewTextFormatter(formatter.TextFormatterConfig{
		Language: cfg.Output.Language,
		Verbose:  cfg.Output.Verbose,
	}))
	formatterRegistry.MustRegister(formatter.NewJSONFormatter(db))

	selectedFormatter, ok := formatterRegistry.Get(cfg.Output.Format)
	if !ok {
		return nil, fmt.Errorf("formatter %q is not registered", cfg.Output.Format)
	}

	return &appContainer{
		config:    cfg,
		cardDB:    db,
		analyzer:  logAnalyzer,
		formatter: selectedFormatter,
	}, nil
}

func loadCLIConfig(configPath string) (*config.Config, error) {
	if strings.TrimSpace(configPath) != "" {
		loader, err := config.NewLoader()
		if err != nil {
			return nil, fmt.Errorf("create config loader: %w", err)
		}

		cfg, err := loader.Load(configPath)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	cfg := config.DefaultConfig()
	cfg.LogsPath = resolveCLIPath(workingDir, cfg.LogsPath)
	cfg.CardsDBPath = resolveCLIPath(workingDir, cfg.CardsDBPath)
	if cfg.Output.Destination == "file" && cfg.Output.FilePath != "" {
		cfg.Output.FilePath = resolveCLIPath(workingDir, cfg.Output.FilePath)
	}

	return &cfg, nil
}

func applyAnalyzeOverrides(cfg *config.Config, options analyzeOptions) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if strings.TrimSpace(options.logsPath) != "" {
		cfg.LogsPath = resolveCLIPath(workingDir, options.logsPath)
	}
	if strings.TrimSpace(options.cardsDBPath) != "" {
		cfg.CardsDBPath = resolveCLIPath(workingDir, options.cardsDBPath)
	}
	if strings.TrimSpace(options.format) != "" {
		cfg.Output.Format = options.format
	}
	if strings.TrimSpace(options.output) != "" {
		cfg.Output.Destination = options.output
	}
	if strings.TrimSpace(options.outputFile) != "" {
		cfg.Output.FilePath = resolveCLIPath(workingDir, options.outputFile)
	}
	if strings.TrimSpace(options.language) != "" {
		cfg.Output.Language = options.language
	}
	if options.verbose {
		cfg.Output.Verbose = true
	}

	return nil
}

func validateCLIConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if cfg.Output.Format != "text" && cfg.Output.Format != "json" {
		return fmt.Errorf("unsupported output format %q", cfg.Output.Format)
	}
	switch cfg.Output.Language {
	case "ru", "en":
	default:
		return fmt.Errorf("unsupported language %q", cfg.Output.Language)
	}
	switch cfg.Output.Destination {
	case "stdout":
	case "file":
		if strings.TrimSpace(cfg.Output.FilePath) == "" {
			return fmt.Errorf("output file path is required when destination=file")
		}
	default:
		return fmt.Errorf("unsupported output destination %q", cfg.Output.Destination)
	}

	if err := validateDirectory(cfg.LogsPath); err != nil {
		return fmt.Errorf("invalid logs path: %w", err)
	}
	if err := validateFile(cfg.CardsDBPath); err != nil {
		return fmt.Errorf("invalid cards database path: %w", err)
	}

	return nil
}

func validateDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	return nil
}

func validateFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	return nil
}

func resolveCLIPath(baseDir, path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(baseDir, path))
}
