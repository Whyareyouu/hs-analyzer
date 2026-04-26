package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	if cfg.LogsPath != DefaultLogsPath {
		t.Fatalf("unexpected logs path: %s", cfg.LogsPath)
	}
	if cfg.CardsDBPath != DefaultCardsDBPath {
		t.Fatalf("unexpected cards db path: %s", cfg.CardsDBPath)
	}
	if cfg.Output.Format != DefaultOutputFormat {
		t.Fatalf("unexpected output format: %s", cfg.Output.Format)
	}
	if cfg.Output.Destination != DefaultOutputDestination {
		t.Fatalf("unexpected output destination: %s", cfg.Output.Destination)
	}
	if cfg.Output.JSONMode != DefaultJSONMode {
		t.Fatalf("unexpected output json mode: %s", cfg.Output.JSONMode)
	}
	if cfg.Output.Language != DefaultOutputLanguage {
		t.Fatalf("unexpected output language: %s", cfg.Output.Language)
	}
	if len(cfg.Filters.GameMode) != 1 || cfg.Filters.GameMode[0] != "standard" {
		t.Fatalf("unexpected game modes: %#v", cfg.Filters.GameMode)
	}
	if cfg.Performance.Workers != DefaultPerformanceWorker {
		t.Fatalf("unexpected workers: %d", cfg.Performance.Workers)
	}
	if cfg.Performance.BufferSize != DefaultBufferSize {
		t.Fatalf("unexpected buffer size: %d", cfg.Performance.BufferSize)
	}
}

func TestLoaderLoadDefaults(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	logsDir := filepath.Join(workspace, "logs")
	dataDir := filepath.Join(workspace, "data")
	cardsPath := filepath.Join(dataDir, "cards.json")

	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatalf("create logs dir: %v", err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}
	if err := os.WriteFile(cardsPath, []byte("[]"), 0o600); err != nil {
		t.Fatalf("write cards db: %v", err)
	}

	loader := &Loader{workingDir: workspace}
	cfg, err := loader.Load("")
	if err != nil {
		t.Fatalf("load defaults: %v", err)
	}

	if cfg.LogsPath != logsDir {
		t.Fatalf("unexpected resolved logs path: %s", cfg.LogsPath)
	}
	if cfg.CardsDBPath != cardsPath {
		t.Fatalf("unexpected resolved cards path: %s", cfg.CardsDBPath)
	}
}

func TestLoaderLoadFromYAML(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	configDir := filepath.Join(workspace, "configs")
	logsDir := filepath.Join(workspace, "custom-logs")
	dataDir := filepath.Join(workspace, "assets")
	outputDir := filepath.Join(workspace, "output")
	cardsPath := filepath.Join(dataDir, "cards.json")
	configPath := filepath.Join(configDir, "config.yaml")

	for _, dir := range []string{configDir, logsDir, dataDir, outputDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create dir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(cardsPath, []byte("[]"), 0o600); err != nil {
		t.Fatalf("write cards db: %v", err)
	}

	yaml := strings.TrimSpace(`
logs_path: ../custom-logs
cards_db_path: ../assets/cards.json
output:
  format: json
  json_mode: human
  destination: file
  file_path: ../output/result.txt
  language: en
  verbose: true
filters:
  game_mode:
    - standard
    - ranked
  min_turns: 3
  max_turns: 12
performance:
  workers: 4
  buffer_size: 8192
`)
	if err := os.WriteFile(configPath, []byte(yaml), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loader := &Loader{workingDir: workspace}
	cfg, err := loader.Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.LogsPath != logsDir {
		t.Fatalf("unexpected logs path: %s", cfg.LogsPath)
	}
	if cfg.CardsDBPath != cardsPath {
		t.Fatalf("unexpected cards path: %s", cfg.CardsDBPath)
	}
	if cfg.Output.FilePath != filepath.Join(outputDir, "result.txt") {
		t.Fatalf("unexpected output file path: %s", cfg.Output.FilePath)
	}
	if cfg.Output.Format != "json" {
		t.Fatalf("unexpected output format: %s", cfg.Output.Format)
	}
	if cfg.Output.JSONMode != "human" {
		t.Fatalf("unexpected json mode: %s", cfg.Output.JSONMode)
	}
	if !cfg.Output.Verbose {
		t.Fatalf("expected verbose output")
	}
	if cfg.Performance.Workers != 4 {
		t.Fatalf("unexpected workers: %d", cfg.Performance.Workers)
	}
}

func TestLoaderLoadFromYAMLWithExplicitPowerLogPath(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	configDir := filepath.Join(workspace, "configs")
	logsDir := filepath.Join(workspace, "custom-logs")
	dataDir := filepath.Join(workspace, "assets")
	cardsPath := filepath.Join(dataDir, "cards.json")
	configPath := filepath.Join(configDir, "config.yaml")
	powerLogPath := filepath.Join(logsDir, "Power.log")

	for _, dir := range []string{configDir, logsDir, dataDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create dir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(cardsPath, []byte("[]"), 0o600); err != nil {
		t.Fatalf("write cards db: %v", err)
	}
	if err := os.WriteFile(powerLogPath, []byte("power"), 0o600); err != nil {
		t.Fatalf("write power log: %v", err)
	}

	yaml := strings.TrimSpace(`
logs_path: ../custom-logs/Power.log
cards_db_path: ../assets/cards.json
`)
	if err := os.WriteFile(configPath, []byte(yaml), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loader := &Loader{workingDir: workspace}
	cfg, err := loader.Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.LogsPath != logsDir {
		t.Fatalf("unexpected logs dir: %s", cfg.LogsPath)
	}
	if cfg.PowerLogPath != powerLogPath {
		t.Fatalf("unexpected power log path: %s", cfg.PowerLogPath)
	}
}

func TestLoaderLoadRejectsInvalidPaths(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	configPath := filepath.Join(workspace, "config.yaml")

	yaml := strings.TrimSpace(`
logs_path: ./missing-logs
cards_db_path: ./missing-cards.json
`)
	if err := os.WriteFile(configPath, []byte(yaml), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loader := &Loader{workingDir: workspace}
	_, err := loader.Load(configPath)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid logs path") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoaderLoadRejectsInvalidOutputConfig(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	logsDir := filepath.Join(workspace, "logs")
	dataDir := filepath.Join(workspace, "data")
	configPath := filepath.Join(workspace, "config.yaml")

	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatalf("create logs dir: %v", err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "cards.json"), []byte("[]"), 0o600); err != nil {
		t.Fatalf("write cards db: %v", err)
	}

	yaml := strings.TrimSpace(`
logs_path: ./logs
cards_db_path: ./data/cards.json
output:
  format: xml
  destination: stdout
  language: ru
`)
	if err := os.WriteFile(configPath, []byte(yaml), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loader := &Loader{workingDir: workspace}
	_, err := loader.Load(configPath)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid output config") {
		t.Fatalf("unexpected error: %v", err)
	}
}
