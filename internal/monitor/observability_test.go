package monitor

import (
	"errors"
	"strings"
	"testing"
)

func TestHealthTrackerDefaultHealthy(t *testing.T) {
	t.Parallel()

	tracker := NewHealthTracker()
	snapshot := tracker.Snapshot()
	if snapshot.Watcher != WatcherHealthHealthy {
		t.Fatalf("unexpected watcher health: %s", snapshot.Watcher)
	}
	if snapshot.Storage != StorageHealthHealthy {
		t.Fatalf("unexpected storage health: %s", snapshot.Storage)
	}
	if snapshot.SessionPhase != SessionPhaseIdle {
		t.Fatalf("unexpected session phase: %s", snapshot.SessionPhase)
	}

	check := tracker.SelfCheck()
	if !check.Healthy {
		t.Fatalf("expected healthy self-check, got issues: %#v", check.Issues)
	}
}

func TestHealthTrackerPathMissingTransitions(t *testing.T) {
	t.Parallel()

	tracker := NewHealthTracker()
	tracker.SetPathMissing(true)

	snapshot := tracker.Snapshot()
	if snapshot.Watcher != WatcherHealthPathMissing {
		t.Fatalf("unexpected watcher health after missing path: %s", snapshot.Watcher)
	}
	if !snapshot.PathMissing {
		t.Fatalf("expected path missing flag")
	}

	check := tracker.SelfCheck()
	if check.Healthy {
		t.Fatalf("expected unhealthy self-check")
	}
	if len(check.Issues) == 0 || !strings.Contains(check.Issues[0], "log path missing") {
		t.Fatalf("unexpected self-check issues: %#v", check.Issues)
	}

	tracker.SetPathMissing(false)
	tracker.SetWatcherHealthy()

	snapshot = tracker.Snapshot()
	if snapshot.Watcher != WatcherHealthHealthy {
		t.Fatalf("unexpected watcher health after recovery: %s", snapshot.Watcher)
	}
	if snapshot.PathMissing {
		t.Fatalf("unexpected path missing flag after recovery")
	}
}

func TestHealthTrackerStorageDegradedTransitions(t *testing.T) {
	t.Parallel()

	tracker := NewHealthTracker()
	tracker.SetStorageError(errors.New("storage unavailable"))

	snapshot := tracker.Snapshot()
	if snapshot.Storage != StorageHealthDegraded {
		t.Fatalf("unexpected storage health: %s", snapshot.Storage)
	}
	if !strings.Contains(snapshot.LastStorageError, "storage unavailable") {
		t.Fatalf("unexpected storage error message: %q", snapshot.LastStorageError)
	}

	check := tracker.SelfCheck()
	if check.Healthy {
		t.Fatalf("expected unhealthy self-check")
	}
	if len(check.Issues) == 0 || !strings.Contains(check.Issues[0], "storage degraded") {
		t.Fatalf("unexpected self-check issues: %#v", check.Issues)
	}

	tracker.SetStorageHealthy()
	snapshot = tracker.Snapshot()
	if snapshot.Storage != StorageHealthHealthy {
		t.Fatalf("unexpected storage health after recovery: %s", snapshot.Storage)
	}
}

func TestHealthTrackerSessionState(t *testing.T) {
	t.Parallel()

	tracker := NewHealthTracker()
	tracker.SetSessionState(DetectorState{
		Phase:            SessionPhaseActiveMatch,
		PendingCount:     1,
		ActiveEventCount: 4,
	})

	snapshot := tracker.Snapshot()
	if snapshot.SessionPhase != SessionPhaseActiveMatch {
		t.Fatalf("unexpected session phase: %s", snapshot.SessionPhase)
	}
	if !snapshot.SessionActive {
		t.Fatalf("expected session active flag")
	}
}
