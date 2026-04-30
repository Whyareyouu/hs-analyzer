package main

import (
	"os"
	"path/filepath"
	"testing"

	"hsanalyzer/internal/monitor"
)

func TestMonitorRuntimeStateRoundTrip(t *testing.T) {
	t.Parallel()

	statePath := filepath.Join(t.TempDir(), "monitor_state.json")
	want := monitorRuntimeState{
		Version: monitorStateVersion,
		Readers: map[string]monitorReaderCheckpoint{
			"power": {
				Path:   filepath.Clean("C:\\Logs\\Power.log"),
				Offset: 128,
			},
		},
		LiveCardStats: &monitor.LiveCardStatsSnapshot{
			SchemaVersion:     monitor.LiveCardStatsSchemaVersion,
			ProjectionVersion: 2,
			RecoveryStatus:    monitor.RecoveryStatusFresh,
			Match: monitor.LiveCardStatsMatchSnapshot{
				Key:    "game:test",
				Status: monitor.LiveCardStatsMatchStatusActive,
			},
			Players: monitor.LiveCardStatsPlayersSnapshot{
				Self: monitor.LiveCardStatsSideSnapshot{
					TotalPlays:  1,
					UniqueCards: 1,
					Cards: []monitor.LiveCardStatsCardEntry{
						{CardID: "CARD_1", PlayCount: 1, FirstPlaySeq: 1, LastPlaySeq: 1},
					},
				},
				Opponent: monitor.LiveCardStatsSideSnapshot{
					Cards: make([]monitor.LiveCardStatsCardEntry, 0),
				},
			},
		},
	}

	if err := saveMonitorRuntimeState(statePath, want); err != nil {
		t.Fatalf("save monitor state: %v", err)
	}

	got, err := loadMonitorRuntimeState(statePath)
	if err != nil {
		t.Fatalf("load monitor state: %v", err)
	}

	checkpoint, ok := got.Readers["power"]
	if !ok {
		t.Fatalf("missing power checkpoint")
	}
	if checkpoint.Offset != 128 {
		t.Fatalf("unexpected offset: %d", checkpoint.Offset)
	}
	if filepath.Clean(checkpoint.Path) != filepath.Clean("C:\\Logs\\Power.log") {
		t.Fatalf("unexpected checkpoint path: %s", checkpoint.Path)
	}
	if got.LiveCardStats == nil {
		t.Fatalf("expected live card stats snapshot in runtime state")
	}
	if got.LiveCardStats.Match.Key != "game:test" {
		t.Fatalf("unexpected live card stats match key: %q", got.LiveCardStats.Match.Key)
	}
	if len(got.LiveCardStats.Players.Self.Cards) != 1 || got.LiveCardStats.Players.Self.Cards[0].CardID != "CARD_1" {
		t.Fatalf("unexpected live card stats snapshot: %#v", got.LiveCardStats)
	}
}

func TestLoadMonitorRuntimeStateMissingFile(t *testing.T) {
	t.Parallel()

	statePath := filepath.Join(t.TempDir(), "missing.json")
	state, err := loadMonitorRuntimeState(statePath)
	if err != nil {
		t.Fatalf("load missing state: %v", err)
	}
	if state.Version != monitorStateVersion {
		t.Fatalf("unexpected version: %d", state.Version)
	}
	if len(state.Readers) != 0 {
		t.Fatalf("unexpected readers: %#v", state.Readers)
	}
}

func TestResolveReaderStartOffset(t *testing.T) {
	t.Parallel()

	state := monitorRuntimeState{
		Readers: map[string]monitorReaderCheckpoint{
			"power": {
				Path:   filepath.Clean("D:\\logs\\Power.log"),
				Offset: 777,
			},
		},
	}

	offset := resolveReaderStartOffset(state, "power", filepath.Clean("D:\\logs\\Power.log"))
	if offset != 777 {
		t.Fatalf("unexpected resolved offset: %d", offset)
	}

	mismatch := resolveReaderStartOffset(state, "power", filepath.Clean("D:\\logs\\Power_old.log"))
	if mismatch != 0 {
		t.Fatalf("expected zero offset for path mismatch, got: %d", mismatch)
	}
}

func TestResolveLiveReaderStartOffsetStartsFromBeginningWithoutCheckpoint(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "Power.log")
	if err := os.WriteFile(logPath, []byte("power"), 0o600); err != nil {
		t.Fatalf("write log file: %v", err)
	}

	offset, err := resolveLiveReaderStartOffset(monitorRuntimeState{}, "power", logPath)
	if err != nil {
		t.Fatalf("resolve live reader offset: %v", err)
	}
	if offset != 0 {
		t.Fatalf("expected bootstrap offset 0, got %d", offset)
	}
}
