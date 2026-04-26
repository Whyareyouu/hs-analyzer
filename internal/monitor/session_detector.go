package monitor

import (
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/types"
)

// SessionPhase reflects whether the detector currently tracks an active match.
type SessionPhase string

const (
	SessionPhaseIdle        SessionPhase = "idle"
	SessionPhaseActiveMatch SessionPhase = "active_match"
)

// CompletionReason explains why a session was emitted.
type CompletionReason string

const (
	CompletionReasonGameEnd       CompletionReason = "game_end"
	CompletionReasonNextGameStart CompletionReason = "next_game_start"
	CompletionReasonFlush         CompletionReason = "flush"
	CompletionReasonStreamEnd     CompletionReason = "stream_end"
)

// DetectorState exposes the current state of the session detector.
type DetectorState struct {
	Phase            SessionPhase
	PendingCount     int
	ActiveEventCount int
}

// CompletedSession contains the events emitted for one detected match/session.
type CompletedSession struct {
	Events        []game.Event
	Reason        CompletionReason
	Complete      bool
	HasGameStart  bool
	HasGameEnd    bool
	StartedByLive bool
}

// SessionDetector incrementally groups a live event stream into sessions.
//
// It is designed for live monitoring where events arrive over time:
// pre-game events are buffered until GameStart, active sessions stay open while
// new events keep arriving, and a session is completed on GameEnd, on the next
// GameStart, or on explicit Flush/Close by the caller.
type SessionDetector struct {
	pending []game.Event
	active  []game.Event
}

// NewSessionDetector creates an empty session detector.
func NewSessionDetector() *SessionDetector {
	return &SessionDetector{
		pending: make([]game.Event, 0),
		active:  make([]game.Event, 0),
	}
}

// State returns the detector's current state.
func (d *SessionDetector) State() DetectorState {
	if d == nil {
		return DetectorState{}
	}

	phase := SessionPhaseIdle
	if len(d.active) > 0 {
		phase = SessionPhaseActiveMatch
	}

	return DetectorState{
		Phase:            phase,
		PendingCount:     len(d.pending),
		ActiveEventCount: len(d.active),
	}
}

// Process accepts one event from the live stream and may emit completed sessions.
func (d *SessionDetector) Process(event game.Event) []CompletedSession {
	if d == nil || event == nil {
		return nil
	}

	switch event.Type() {
	case types.EventTypeGameStart:
		emitted := make([]CompletedSession, 0, 1)
		if len(d.active) > 0 {
			emitted = append(emitted, d.emitActive(CompletionReasonNextGameStart, false))
		}

		d.active = append(d.active[:0], d.pending...)
		d.pending = d.pending[:0]
		d.active = append(d.active, event)
		return emitted

	case types.EventTypeGameEnd:
		if len(d.active) == 0 {
			d.active = append(d.active[:0], d.pending...)
			d.pending = d.pending[:0]
		}
		d.active = append(d.active, event)
		return []CompletedSession{d.emitActive(CompletionReasonGameEnd, true)}

	default:
		if len(d.active) > 0 {
			d.active = append(d.active, event)
			return nil
		}

		d.pending = append(d.pending, event)
		return nil
	}
}

// Flush emits the currently buffered session, if any, as incomplete unless it
// was already closed by GameEnd.
func (d *SessionDetector) Flush(reason CompletionReason) (CompletedSession, bool) {
	if d == nil {
		return CompletedSession{}, false
	}

	switch {
	case len(d.active) > 0:
		return d.emitActive(reason, false), true
	case len(d.pending) > 0:
		session := CompletedSession{
			Events:        cloneEvents(d.pending),
			Reason:        reason,
			Complete:      false,
			HasGameStart:  containsEventType(d.pending, types.EventTypeGameStart),
			HasGameEnd:    containsEventType(d.pending, types.EventTypeGameEnd),
			StartedByLive: false,
		}
		d.pending = d.pending[:0]
		return session, true
	default:
		return CompletedSession{}, false
	}
}

func (d *SessionDetector) emitActive(reason CompletionReason, complete bool) CompletedSession {
	session := CompletedSession{
		Events:        cloneEvents(d.active),
		Reason:        reason,
		Complete:      complete,
		HasGameStart:  containsEventType(d.active, types.EventTypeGameStart),
		HasGameEnd:    containsEventType(d.active, types.EventTypeGameEnd),
		StartedByLive: true,
	}

	d.active = d.active[:0]
	return session
}

func cloneEvents(events []game.Event) []game.Event {
	return append([]game.Event(nil), events...)
}

func containsEventType(events []game.Event, eventType types.EventType) bool {
	for _, event := range events {
		if event != nil && event.Type() == eventType {
			return true
		}
	}
	return false
}
