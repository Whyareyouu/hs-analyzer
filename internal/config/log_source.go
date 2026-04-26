package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const activePowerLogName = "Power.log"

// LogSourceOptions controls how log locations are resolved and validated.
type LogSourceOptions struct {
	LogsPathExplicit     bool
	PowerLogPathExplicit bool
	AllowAutoDiscovery   bool
}

// NormalizeLogSource resolves logs_path / power_log_path into a consistent state.
func NormalizeLogSource(cfg *Config, options LogSourceOptions) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	cfg.LogsPath = strings.TrimSpace(cfg.LogsPath)
	cfg.PowerLogPath = strings.TrimSpace(cfg.PowerLogPath)

	if cfg.PowerLogPath != "" {
		if err := validatePowerLogFile(cfg.PowerLogPath); err != nil {
			return err
		}
		cfg.PowerLogPath = filepath.Clean(cfg.PowerLogPath)
		cfg.LogsPath = filepath.Dir(cfg.PowerLogPath)
		return nil
	}

	if cfg.LogsPath != "" {
		resolved, err := normalizeLogsPath(cfg.LogsPath)
		if err == nil {
			cfg.LogsPath = resolved.LogsDir
			cfg.PowerLogPath = resolved.PowerLogPath
			return nil
		}

		if !options.AllowAutoDiscovery || options.LogsPathExplicit {
			return err
		}
	}

	if !options.AllowAutoDiscovery {
		if cfg.LogsPath == "" {
			return errors.New("path is empty")
		}
		return nil
	}

	discovered, ok := DiscoverHearthstoneLogsDir()
	if !ok {
		if cfg.LogsPath != "" {
			return fmt.Errorf("path %q not found and auto-discovery did not find Hearthstone logs", cfg.LogsPath)
		}
		return errors.New("hearthstone logs path was not provided and auto-discovery did not find Hearthstone logs")
	}

	cfg.LogsPath = discovered
	cfg.PowerLogPath = filepath.Join(discovered, activePowerLogName)
	return nil
}

type normalizedLogSource struct {
	LogsDir      string
	PowerLogPath string
}

func normalizeLogsPath(path string) (normalizedLogSource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return normalizedLogSource{}, err
	}

	cleaned := filepath.Clean(path)
	if info.IsDir() {
		return normalizedLogSource{LogsDir: cleaned}, nil
	}

	if !isActivePowerLogPath(cleaned) {
		return normalizedLogSource{}, fmt.Errorf("%s is not %s", cleaned, activePowerLogName)
	}

	return normalizedLogSource{
		LogsDir:      filepath.Dir(cleaned),
		PowerLogPath: cleaned,
	}, nil
}

func validatePowerLogFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	if !isActivePowerLogPath(path) {
		return fmt.Errorf("%s is not %s", path, activePowerLogName)
	}
	return nil
}

func isActivePowerLogPath(path string) bool {
	return strings.EqualFold(filepath.Base(filepath.Clean(path)), activePowerLogName)
}

// DiscoverHearthstoneLogsDir tries to find the standard Hearthstone Logs directory on Windows.
func DiscoverHearthstoneLogsDir() (string, bool) {
	candidates := make([]string, 0, 2)

	if localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); localAppData != "" {
		candidates = append(candidates, filepath.Join(localAppData, "Blizzard", "Hearthstone", "Logs"))
	}

	if homeDir, err := os.UserHomeDir(); err == nil && strings.TrimSpace(homeDir) != "" {
		candidates = append(candidates, filepath.Join(homeDir, "AppData", "Local", "Blizzard", "Hearthstone", "Logs"))
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}

		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			powerPath := filepath.Join(candidate, activePowerLogName)
			if powerInfo, powerErr := os.Stat(powerPath); powerErr == nil && !powerInfo.IsDir() {
				return candidate, true
			}
		}
	}

	return "", false
}
