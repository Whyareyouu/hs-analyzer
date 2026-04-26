package monitor

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// WatcherHealth describes health of live log watching.
type WatcherHealth string

const (
	WatcherHealthHealthy     WatcherHealth = "watcher_healthy"
	WatcherHealthPathMissing WatcherHealth = "log_path_missing"
	WatcherHealthError       WatcherHealth = "watcher_error"
)

// StorageHealth describes health of persistence layer.
type StorageHealth string

const (
	StorageHealthHealthy  StorageHealth = "storage_healthy"
	StorageHealthDegraded StorageHealth = "storage_degraded"
)

// LiveHealthSnapshot is a compact diagnostics snapshot for monitor mode.
type LiveHealthSnapshot struct {
	Watcher          WatcherHealth
	Storage          StorageHealth
	SessionPhase     SessionPhase
	SessionActive    bool
	PathMissing      bool
	LastWatcherError string
	LastStorageError string
	UpdatedAt        time.Time
}

// SelfCheckResult contains minimal self-diagnostics output.
type SelfCheckResult struct {
	Healthy bool
	Issues  []string
}

// HealthTracker tracks observability and health transitions for live mode.
type HealthTracker struct {
	mu       sync.Mutex
	snapshot LiveHealthSnapshot
}

// NewHealthTracker creates a tracker with healthy defaults.
func NewHealthTracker() *HealthTracker {
	return &HealthTracker{
		snapshot: LiveHealthSnapshot{
			Watcher:      WatcherHealthHealthy,
			Storage:      StorageHealthHealthy,
			SessionPhase: SessionPhaseIdle,
			UpdatedAt:    time.Now(),
		},
	}
}

// SetSessionState updates session diagnostics from detector state.
func (t *HealthTracker) SetSessionState(state DetectorState) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	t.snapshot.SessionPhase = state.Phase
	t.snapshot.SessionActive = state.Phase == SessionPhaseActiveMatch
	t.snapshot.UpdatedAt = time.Now()
}

// SetPathMissing marks watcher health based on live log path availability.
func (t *HealthTracker) SetPathMissing(missing bool) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	t.snapshot.PathMissing = missing
	if missing {
		t.snapshot.Watcher = WatcherHealthPathMissing
	} else if t.snapshot.Watcher == WatcherHealthPathMissing {
		t.snapshot.Watcher = WatcherHealthHealthy
	}
	t.snapshot.UpdatedAt = time.Now()
}

// SetWatcherError marks watcher as degraded with the latest error details.
func (t *HealthTracker) SetWatcherError(err error) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	t.snapshot.Watcher = WatcherHealthError
	t.snapshot.LastWatcherError = strings.TrimSpace(errorString(err))
	t.snapshot.UpdatedAt = time.Now()
}

// SetWatcherHealthy marks watcher as healthy and clears last error.
func (t *HealthTracker) SetWatcherHealthy() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.snapshot.PathMissing {
		t.snapshot.Watcher = WatcherHealthPathMissing
	} else {
		t.snapshot.Watcher = WatcherHealthHealthy
	}
	t.snapshot.LastWatcherError = ""
	t.snapshot.UpdatedAt = time.Now()
}

// SetStorageError marks storage as degraded with last persistence error.
func (t *HealthTracker) SetStorageError(err error) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	t.snapshot.Storage = StorageHealthDegraded
	t.snapshot.LastStorageError = strings.TrimSpace(errorString(err))
	t.snapshot.UpdatedAt = time.Now()
}

// SetStorageHealthy marks storage as healthy and clears last storage error.
func (t *HealthTracker) SetStorageHealthy() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	t.snapshot.Storage = StorageHealthHealthy
	t.snapshot.LastStorageError = ""
	t.snapshot.UpdatedAt = time.Now()
}

// Snapshot returns the current diagnostics state.
func (t *HealthTracker) Snapshot() LiveHealthSnapshot {
	if t == nil {
		return LiveHealthSnapshot{}
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	return t.snapshot
}

// SelfCheck reports current health and a compact issue list.
func (t *HealthTracker) SelfCheck() SelfCheckResult {
	snapshot := t.Snapshot()
	issues := make([]string, 0, 3)

	if snapshot.Watcher == WatcherHealthPathMissing {
		issues = append(issues, "log path missing")
	}
	if snapshot.Watcher == WatcherHealthError {
		if snapshot.LastWatcherError != "" {
			issues = append(issues, fmt.Sprintf("watcher error: %s", snapshot.LastWatcherError))
		} else {
			issues = append(issues, "watcher error")
		}
	}
	if snapshot.Storage == StorageHealthDegraded {
		if snapshot.LastStorageError != "" {
			issues = append(issues, fmt.Sprintf("storage degraded: %s", snapshot.LastStorageError))
		} else {
			issues = append(issues, "storage degraded")
		}
	}

	return SelfCheckResult{
		Healthy: len(issues) == 0,
		Issues:  issues,
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
