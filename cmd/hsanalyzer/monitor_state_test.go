package main

import (
	"path/filepath"
	"testing"
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
