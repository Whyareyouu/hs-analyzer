package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	candidates := make([]string, 0, 8)

	if localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); localAppData != "" {
		candidates = append(candidates, filepath.Join(localAppData, "Blizzard", "Hearthstone", "Logs"))
	}

	if homeDir, err := os.UserHomeDir(); err == nil && strings.TrimSpace(homeDir) != "" {
		candidates = append(candidates, filepath.Join(homeDir, "AppData", "Local", "Blizzard", "Hearthstone", "Logs"))
	}

	for _, root := range discoverWindowsDriveRoots() {
		candidates = append(candidates,
			filepath.Join(root, "Hearthstone", "Logs"),
			filepath.Join(root, "Games", "Hearthstone", "Logs"),
			filepath.Join(root, "Blizzard", "Hearthstone", "Logs"),
		)
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}

		if resolved, ok := resolveHearthstoneLogsCandidate(candidate); ok {
			return resolved, true
		}
	}

	return "", false
}

func discoverWindowsDriveRoots() []string {
	roots := make([]string, 0, 4)

	for letter := 'A'; letter <= 'Z'; letter++ {
		root := fmt.Sprintf("%c:\\", letter)
		if info, err := os.Stat(root); err == nil && info.IsDir() {
			roots = append(roots, root)
		}
	}

	return roots
}

func resolveHearthstoneLogsCandidate(candidate string) (string, bool) {
	info, err := os.Stat(candidate)
	if err != nil || !info.IsDir() {
		return "", false
	}

	powerPath := filepath.Join(candidate, activePowerLogName)
	if powerInfo, powerErr := os.Stat(powerPath); powerErr == nil && !powerInfo.IsDir() {
		return candidate, true
	}

	children, err := os.ReadDir(candidate)
	if err != nil {
		return "", false
	}

	type nestedCandidate struct {
		path    string
		modTime int64
	}

	nested := make([]nestedCandidate, 0, len(children))
	for _, child := range children {
		if !child.IsDir() {
			continue
		}

		childPath := filepath.Join(candidate, child.Name())
		powerPath := filepath.Join(childPath, activePowerLogName)
		powerInfo, err := os.Stat(powerPath)
		if err != nil || powerInfo.IsDir() {
			continue
		}

		nested = append(nested, nestedCandidate{
			path:    childPath,
			modTime: powerInfo.ModTime().UnixNano(),
		})
	}

	sort.Slice(nested, func(i, j int) bool {
		if nested[i].modTime == nested[j].modTime {
			return nested[i].path > nested[j].path
		}
		return nested[i].modTime > nested[j].modTime
	})

	if len(nested) == 0 {
		return "", false
	}

	return nested[0].path, true
}
