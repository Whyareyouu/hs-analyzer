package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"hsanalyzer/internal/monitor"
	"hsanalyzer/pkg/contracts/overlay"
)

const monitorStateVersion = 1

type monitorReaderCheckpoint struct {
	Path      string `json:"path"`
	Offset    int64  `json:"offset"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type monitorRuntimeState struct {
	Version       int                                `json:"version"`
	UpdatedAt     string                             `json:"updated_at,omitempty"`
	Readers       map[string]monitorReaderCheckpoint `json:"readers,omitempty"`
	LiveCardStats *monitor.LiveCardStatsSnapshot     `json:"live_card_stats,omitempty"`
	Overlay       *overlay.Snapshot                  `json:"overlay,omitempty"`
}

func defaultMonitorStatePath() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	projectRoot, err := findProjectRoot(workingDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, "output", "live_monitor_state.json"), nil
}

func loadMonitorRuntimeState(path string) (monitorRuntimeState, error) {
	if path == "" {
		return monitorRuntimeState{}, errors.New("state path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return monitorRuntimeState{
				Version: monitorStateVersion,
				Readers: make(map[string]monitorReaderCheckpoint),
			}, nil
		}
		return monitorRuntimeState{}, fmt.Errorf("read monitor state file: %w", err)
	}

	var state monitorRuntimeState
	if err := json.Unmarshal(data, &state); err != nil {
		return monitorRuntimeState{}, fmt.Errorf("parse monitor state file: %w", err)
	}

	if state.Readers == nil {
		state.Readers = make(map[string]monitorReaderCheckpoint)
	}
	if state.Version == 0 {
		state.Version = monitorStateVersion
	}

	return state, nil
}

func saveMonitorRuntimeState(path string, state monitorRuntimeState) error {
	if path == "" {
		return errors.New("state path is empty")
	}
	state.Version = monitorStateVersion
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if state.Readers == nil {
		state.Readers = make(map[string]monitorReaderCheckpoint)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create monitor state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal monitor state: %w", err)
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("write temporary monitor state: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace monitor state file: %w", err)
	}

	return nil
}

func resolveReaderStartOffset(state monitorRuntimeState, source, path string) int64 {
	checkpoint, ok := state.Readers[source]
	if !ok {
		return 0
	}
	if filepath.Clean(checkpoint.Path) != filepath.Clean(path) {
		return 0
	}
	if checkpoint.Offset < 0 {
		return 0
	}
	return checkpoint.Offset
}
