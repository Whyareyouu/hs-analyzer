package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var (
	errInvalidLogsPath    = errors.New("invalid logs path")
	errInvalidCardsDBPath = errors.New("invalid cards database path")
	errInvalidOutput      = errors.New("invalid output config")
	errInvalidFilters     = errors.New("invalid filters config")
	errInvalidPerformance = errors.New("invalid performance config")
)

// Loader загружает конфигурацию приложения из YAML.
type Loader struct {
	workingDir string
}

// NewLoader создает новый загрузчик конфигурации.
func NewLoader() (*Loader, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	return &Loader{workingDir: workingDir}, nil
}

// Load загружает конфигурацию из файла или использует значения по умолчанию.
func (l *Loader) Load(configPath string) (*Config, error) {
	if l == nil {
		return nil, errors.New("config loader is nil")
	}
	if l.workingDir == "" {
		return nil, errors.New("loader working directory is empty")
	}

	cfg := DefaultConfig()
	baseDir := l.workingDir

	if configPath != "" {
		absoluteConfigPath, err := filepath.Abs(configPath)
		if err != nil {
			return nil, fmt.Errorf("resolve config path: %w", err)
		}

		v := viper.New()
		v.SetConfigFile(absoluteConfigPath)
		v.SetConfigType("yaml")

		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}

		if err := v.UnmarshalExact(&cfg); err != nil {
			return nil, fmt.Errorf("unmarshal config: %w", err)
		}

		baseDir = filepath.Dir(absoluteConfigPath)
	}

	l.resolvePaths(&cfg, baseDir)

	if err := l.validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (l *Loader) resolvePaths(cfg *Config, baseDir string) {
	cfg.LogsPath = resolvePath(baseDir, cfg.LogsPath)
	cfg.CardsDBPath = resolvePath(baseDir, cfg.CardsDBPath)

	if cfg.Output.Destination == "file" {
		cfg.Output.FilePath = resolvePath(baseDir, cfg.Output.FilePath)
	}
}

func (l *Loader) validate(cfg *Config) error {
	if err := validateDirectory(cfg.LogsPath); err != nil {
		return fmt.Errorf("%w: %v", errInvalidLogsPath, err)
	}

	if err := validateFile(cfg.CardsDBPath); err != nil {
		return fmt.Errorf("%w: %v", errInvalidCardsDBPath, err)
	}

	if err := validateOutput(cfg.Output); err != nil {
		return fmt.Errorf("%w: %v", errInvalidOutput, err)
	}

	if err := validateFilters(cfg.Filters); err != nil {
		return fmt.Errorf("%w: %v", errInvalidFilters, err)
	}

	if err := validatePerformance(cfg.Performance); err != nil {
		return fmt.Errorf("%w: %v", errInvalidPerformance, err)
	}

	return nil
}

func resolvePath(baseDir, path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if filepath.IsAbs(trimmed) {
		return filepath.Clean(trimmed)
	}
	return filepath.Join(baseDir, trimmed)
}

func validateDirectory(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("path is empty")
	}

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
	if strings.TrimSpace(path) == "" {
		return errors.New("path is empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	return nil
}

func validateOutput(cfg OutputConfig) error {
	switch cfg.Format {
	case "text", "json":
	default:
		return fmt.Errorf("unsupported format %q", cfg.Format)
	}

	switch cfg.Language {
	case "ru", "en":
	default:
		return fmt.Errorf("unsupported language %q", cfg.Language)
	}

	switch cfg.Destination {
	case "stdout":
		return nil
	case "file":
		if strings.TrimSpace(cfg.FilePath) == "" {
			return errors.New("file_path is required for file destination")
		}

		dir := filepath.Dir(cfg.FilePath)
		if err := validateDirectory(dir); err != nil {
			return fmt.Errorf("output directory: %w", err)
		}
	default:
		return fmt.Errorf("unsupported destination %q", cfg.Destination)
	}

	return nil
}

func validateFilters(cfg FilterConfig) error {
	if cfg.MinTurns < 0 {
		return errors.New("min_turns must be greater than or equal to 0")
	}
	if cfg.MaxTurns < 0 {
		return errors.New("max_turns must be greater than or equal to 0")
	}
	if cfg.MaxTurns > 0 && cfg.MaxTurns < cfg.MinTurns {
		return errors.New("max_turns must be greater than or equal to min_turns")
	}

	for _, mode := range cfg.GameMode {
		if strings.TrimSpace(mode) == "" {
			return errors.New("game_mode must not contain empty values")
		}
	}

	return nil
}

func validatePerformance(cfg PerformanceConfig) error {
	if cfg.Workers <= 0 {
		return errors.New("workers must be greater than 0")
	}
	if cfg.BufferSize <= 0 {
		return errors.New("buffer_size must be greater than 0")
	}

	return nil
}
