package config

import (
	"os"
	"path/filepath"
	"testing"
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
