package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNormalizeLogSourceAcceptsExplicitPowerLogFile(t *testing.T) {
	t.Parallel()

	logsDir := t.TempDir()
	powerLogPath := filepath.Join(logsDir, "Power.log")
	if err := os.WriteFile(powerLogPath, []byte("power"), 0o600); err != nil {
		t.Fatalf("write Power.log: %v", err)
	}

	cfg := &Config{LogsPath: powerLogPath}
	if err := NormalizeLogSource(cfg, LogSourceOptions{
		LogsPathExplicit:   true,
		AllowAutoDiscovery: true,
	}); err != nil {
		t.Fatalf("normalize log source: %v", err)
	}

	if cfg.LogsPath != logsDir {
		t.Fatalf("unexpected logs dir: %s", cfg.LogsPath)
	}
	if cfg.PowerLogPath != powerLogPath {
		t.Fatalf("unexpected power log path: %s", cfg.PowerLogPath)
	}
}

func TestNormalizeLogSourceRejectsPowerOldLogAsExplicitFile(t *testing.T) {
	t.Parallel()

	logsDir := t.TempDir()
	powerLogPath := filepath.Join(logsDir, "Power_old.log")
	if err := os.WriteFile(powerLogPath, []byte("power"), 0o600); err != nil {
		t.Fatalf("write Power_old.log: %v", err)
	}

	cfg := &Config{LogsPath: powerLogPath}
	err := NormalizeLogSource(cfg, LogSourceOptions{
		LogsPathExplicit:   true,
		AllowAutoDiscovery: true,
	})
	if err == nil {
		t.Fatalf("expected error for Power_old.log")
	}
}

func TestNormalizeLogSourceAutoDiscoversWindowsLogsPath(t *testing.T) {
	localAppData := t.TempDir()
	logsDir := filepath.Join(localAppData, "Blizzard", "Hearthstone", "Logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatalf("create logs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(logsDir, "Power.log"), []byte("power"), 0o600); err != nil {
		t.Fatalf("write Power.log: %v", err)
	}
	t.Setenv("LOCALAPPDATA", localAppData)

	cfg := &Config{}
	if err := NormalizeLogSource(cfg, LogSourceOptions{
		AllowAutoDiscovery: true,
	}); err != nil {
		t.Fatalf("normalize log source: %v", err)
	}

	if cfg.LogsPath != logsDir {
		t.Fatalf("unexpected discovered logs path: %s", cfg.LogsPath)
	}
	if cfg.PowerLogPath != filepath.Join(logsDir, "Power.log") {
		t.Fatalf("unexpected discovered power log path: %s", cfg.PowerLogPath)
	}
}

func TestNormalizeLogSourceAutoDiscoversLatestNestedSessionDir(t *testing.T) {
	localAppData := t.TempDir()
	logsRoot := filepath.Join(localAppData, "Blizzard", "Hearthstone", "Logs")
	oldLogsDir := filepath.Join(logsRoot, "Hearthstone_2026_04_27_10_45_49")
	newLogsDir := filepath.Join(logsRoot, "Hearthstone_2026_04_27_15_29_49")

	for _, path := range []string{oldLogsDir, newLogsDir} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatalf("create nested logs dir %s: %v", path, err)
		}
	}
	oldPowerPath := filepath.Join(oldLogsDir, "Power.log")
	newPowerPath := filepath.Join(newLogsDir, "Power.log")
	if err := os.WriteFile(oldPowerPath, []byte("old"), 0o600); err != nil {
		t.Fatalf("write old Power.log: %v", err)
	}
	if err := os.WriteFile(newPowerPath, []byte("new"), 0o600); err != nil {
		t.Fatalf("write new Power.log: %v", err)
	}
	if err := os.Chtimes(oldPowerPath, testTime(2026, 4, 27, 10, 45, 49), testTime(2026, 4, 27, 10, 45, 49)); err != nil {
		t.Fatalf("touch old Power.log: %v", err)
	}
	if err := os.Chtimes(newPowerPath, testTime(2026, 4, 27, 15, 29, 49), testTime(2026, 4, 27, 15, 29, 49)); err != nil {
		t.Fatalf("touch new Power.log: %v", err)
	}
	t.Setenv("LOCALAPPDATA", localAppData)

	cfg := &Config{}
	if err := NormalizeLogSource(cfg, LogSourceOptions{
		AllowAutoDiscovery: true,
	}); err != nil {
		t.Fatalf("normalize log source: %v", err)
	}

	if cfg.LogsPath != newLogsDir {
		t.Fatalf("unexpected discovered nested logs path: %s", cfg.LogsPath)
	}
	if cfg.PowerLogPath != filepath.Join(newLogsDir, "Power.log") {
		t.Fatalf("unexpected discovered nested power log path: %s", cfg.PowerLogPath)
	}
}

func testTime(year int, month time.Month, day, hour, minute, second int) time.Time {
	return time.Date(year, month, day, hour, minute, second, 0, time.UTC)
}
