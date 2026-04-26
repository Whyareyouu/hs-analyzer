package monitor

import (
	"testing"
	"time"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

func TestSessionDetectorStartsSessionOnGameStartAndIncludesPrelude(t *testing.T) {
	t.Parallel()

	detector := NewSessionDetector()
	prelude := testMonitorEvent{eventType: types.EventTypeDeckList, timestamp: time.Unix(1, 0)}
	start := game.NewGameStartEvent("game-1", state.GameModeStandard)

	if emitted := detector.Process(prelude); len(emitted) != 0 {
		t.Fatalf("unexpected emitted sessions before start: %d", len(emitted))
	}

	if emitted := detector.Process(start); len(emitted) != 0 {
		t.Fatalf("unexpected emitted sessions on first start: %d", len(emitted))
	}

	state := detector.State()
	if state.Phase != SessionPhaseActiveMatch {
		t.Fatalf("unexpected phase: %s", state.Phase)
	}
	if state.PendingCount != 0 {
		t.Fatalf("expected prelude to move into active session, pending=%d", state.PendingCount)
	}
	if state.ActiveEventCount != 2 {
		t.Fatalf("unexpected active event count: %d", state.ActiveEventCount)
	}
}

func TestSessionDetectorCompletesSessionOnGameEnd(t *testing.T) {
	t.Parallel()

	detector := NewSessionDetector()
	start := game.NewGameStartEvent("game-1", state.GameModeStandard)
	mid := testMonitorEvent{eventType: types.EventTypeTurnStart, timestamp: time.Unix(2, 0)}
	end := game.NewGameEndEvent(types.PlayerTypePlayer)

	detector.Process(start)
	detector.Process(mid)
	emitted := detector.Process(end)

	if len(emitted) != 1 {
		t.Fatalf("expected one emitted session, got %d", len(emitted))
	}
	if !emitted[0].Complete {
		t.Fatalf("expected completed session")
	}
	if emitted[0].Reason != CompletionReasonGameEnd {
		t.Fatalf("unexpected completion reason: %s", emitted[0].Reason)
	}
	if len(emitted[0].Events) != 3 {
		t.Fatalf("unexpected session event count: %d", len(emitted[0].Events))
	}
	if detector.State().Phase != SessionPhaseIdle {
		t.Fatalf("expected detector to return to idle")
	}
}

func TestSessionDetectorFinalizesPreviousSessionOnNextGameStart(t *testing.T) {
	t.Parallel()

	detector := NewSessionDetector()
	start1 := game.NewGameStartEvent("game-1", state.GameModeStandard)
	turn1 := testMonitorEvent{eventType: types.EventTypeTurnStart, timestamp: time.Unix(2, 0)}
	start2 := game.NewGameStartEvent("game-2", state.GameModeStandard)

	detector.Process(start1)
	detector.Process(turn1)
	emitted := detector.Process(start2)

	if len(emitted) != 1 {
		t.Fatalf("expected one emitted session, got %d", len(emitted))
	}
	if emitted[0].Complete {
		t.Fatalf("session finalized by next start should be incomplete")
	}
	if emitted[0].Reason != CompletionReasonNextGameStart {
		t.Fatalf("unexpected completion reason: %s", emitted[0].Reason)
	}
	if len(emitted[0].Events) != 2 {
		t.Fatalf("unexpected emitted event count: %d", len(emitted[0].Events))
	}
	if detector.State().Phase != SessionPhaseActiveMatch {
		t.Fatalf("expected new active session after second start")
	}
	if detector.State().ActiveEventCount != 1 {
		t.Fatalf("expected second start to seed new active session, got %d events", detector.State().ActiveEventCount)
	}
}

func TestSessionDetectorFlushReturnsIncompleteActiveSession(t *testing.T) {
	t.Parallel()

	detector := NewSessionDetector()
	detector.Process(game.NewGameStartEvent("game-1", state.GameModeStandard))
	detector.Process(testMonitorEvent{eventType: types.EventTypeTurnStart, timestamp: time.Unix(2, 0)})

	session, ok := detector.Flush(CompletionReasonFlush)
	if !ok {
		t.Fatalf("expected flush to emit active session")
	}
	if session.Complete {
		t.Fatalf("expected flushed session to be incomplete")
	}
	if session.Reason != CompletionReasonFlush {
		t.Fatalf("unexpected flush reason: %s", session.Reason)
	}
	if detector.State().Phase != SessionPhaseIdle {
		t.Fatalf("expected detector to become idle after flush")
	}
}

func TestSessionDetectorFlushCanEmitPendingPreludeWithoutStart(t *testing.T) {
	t.Parallel()

	detector := NewSessionDetector()
	detector.Process(testMonitorEvent{eventType: types.EventTypeDeckList, timestamp: time.Unix(1, 0)})
	detector.Process(testMonitorEvent{eventType: types.EventTypeCardDrawn, timestamp: time.Unix(2, 0)})

	session, ok := detector.Flush(CompletionReasonStreamEnd)
	if !ok {
		t.Fatalf("expected flush to emit pending prelude")
	}
	if session.Complete {
		t.Fatalf("pending-only session must be incomplete")
	}
	if session.HasGameStart {
		t.Fatalf("pending-only session should not report game start")
	}
	if len(session.Events) != 2 {
		t.Fatalf("unexpected pending session size: %d", len(session.Events))
	}
}

type testMonitorEvent struct {
	eventType types.EventType
	timestamp time.Time
}

func (e testMonitorEvent) Type() types.EventType {
	return e.eventType
}

func (e testMonitorEvent) Timestamp() time.Time {
	return e.timestamp
}

func (e testMonitorEvent) Apply(*state.State) error {
	return nil
}
