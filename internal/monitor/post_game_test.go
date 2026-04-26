package monitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

type testEventProcessor struct{}

func (testEventProcessor) ProcessAll(_ context.Context, gameState *state.State, events []game.Event) error {
	for _, event := range events {
		if event == nil {
			continue
		}
		if err := event.Apply(gameState); err != nil {
			return err
		}
		gameState.AddEvent(event)
	}
	return nil
}

type stubMatchPersister struct {
	saveFn func(context.Context, PersistedMatch) error
	saved  []PersistedMatch
}

func (p *stubMatchPersister) SaveMatch(ctx context.Context, match PersistedMatch) error {
	if p.saveFn != nil {
		if err := p.saveFn(ctx, match); err != nil {
			return err
		}
	}
	p.saved = append(p.saved, match)
	return nil
}

func newTestPostGamePipeline(t *testing.T, persister MatchPersister, options PostGamePipelineOptions) *PostGamePipeline {
	t.Helper()

	builder, err := NewStateBuilder(testEventProcessor{})
	if err != nil {
		t.Fatalf("new state builder: %v", err)
	}

	postGame, err := NewPostGamePipeline(builder, persister, options)
	if err != nil {
		t.Fatalf("new post-game pipeline: %v", err)
	}

	return postGame
}

func TestPostGamePipelinePersistsCompletedSession(t *testing.T) {
	t.Parallel()

	persister := &stubMatchPersister{}
	postGame := newTestPostGamePipeline(t, persister, PostGamePipelineOptions{})

	session := CompletedSession{
		Events: []game.Event{
			game.NewGameStartEvent("game-1", state.GameModeStandard),
			game.NewTurnStartEvent(types.PlayerTypePlayer),
			game.NewGameEndEvent(types.PlayerTypePlayer),
		},
		Reason:       CompletionReasonGameEnd,
		Complete:     true,
		HasGameStart: true,
		HasGameEnd:   true,
	}

	result, err := postGame.FinalizeSession(context.Background(), session)
	if err != nil {
		t.Fatalf("finalize session: %v", err)
	}
	if !result.Persisted {
		t.Fatalf("expected persisted result")
	}
	if len(persister.saved) != 1 {
		t.Fatalf("expected one saved match, got %d", len(persister.saved))
	}
	if persister.saved[0].State == nil || persister.saved[0].State.GameID != "game-1" {
		t.Fatalf("unexpected persisted state: %#v", persister.saved[0].State)
	}
}

func TestPostGamePipelinePersistsSessionClosedByNextGameStart(t *testing.T) {
	t.Parallel()

	persister := &stubMatchPersister{}
	postGame := newTestPostGamePipeline(t, persister, PostGamePipelineOptions{})

	session := CompletedSession{
		Events: []game.Event{
			game.NewGameStartEvent("game-1", state.GameModeStandard),
			game.NewTurnStartEvent(types.PlayerTypePlayer),
		},
		Reason:       CompletionReasonNextGameStart,
		Complete:     false,
		HasGameStart: true,
	}

	result, err := postGame.FinalizeSession(context.Background(), session)
	if err != nil {
		t.Fatalf("finalize session: %v", err)
	}
	if !result.Persisted {
		t.Fatalf("expected session closed by next game start to be persisted")
	}
}

func TestPostGamePipelineSkipsIncompleteFlushByDefault(t *testing.T) {
	t.Parallel()

	persister := &stubMatchPersister{}
	postGame := newTestPostGamePipeline(t, persister, PostGamePipelineOptions{})

	session := CompletedSession{
		Events: []game.Event{
			game.NewGameStartEvent("game-1", state.GameModeStandard),
			game.NewTurnStartEvent(types.PlayerTypePlayer),
		},
		Reason:       CompletionReasonFlush,
		Complete:     false,
		HasGameStart: true,
	}

	result, err := postGame.FinalizeSession(context.Background(), session)
	if err != nil {
		t.Fatalf("finalize session: %v", err)
	}
	if !result.Skipped {
		t.Fatalf("expected incomplete flush to be skipped by default")
	}
	if len(persister.saved) != 0 {
		t.Fatalf("expected no persisted matches, got %d", len(persister.saved))
	}
}

func TestPostGamePipelineCanPersistIncompleteFlushWhenEnabled(t *testing.T) {
	t.Parallel()

	persister := &stubMatchPersister{}
	postGame := newTestPostGamePipeline(t, persister, PostGamePipelineOptions{PersistIncompleteFlush: true})

	session := CompletedSession{
		Events: []game.Event{
			game.NewGameStartEvent("game-1", state.GameModeStandard),
			game.NewTurnStartEvent(types.PlayerTypePlayer),
		},
		Reason:       CompletionReasonFlush,
		Complete:     false,
		HasGameStart: true,
	}

	result, err := postGame.FinalizeSession(context.Background(), session)
	if err != nil {
		t.Fatalf("finalize session: %v", err)
	}
	if !result.Persisted {
		t.Fatalf("expected incomplete flush to persist when enabled")
	}
}

func TestPostGamePipelinePreventsDuplicateSave(t *testing.T) {
	t.Parallel()

	persister := &stubMatchPersister{}
	postGame := newTestPostGamePipeline(t, persister, PostGamePipelineOptions{})

	session := CompletedSession{
		Events: []game.Event{
			game.NewGameStartEvent("game-dup", state.GameModeStandard),
			game.NewTurnStartEvent(types.PlayerTypePlayer),
			game.NewGameEndEvent(types.PlayerTypePlayer),
		},
		Reason:       CompletionReasonGameEnd,
		Complete:     true,
		HasGameStart: true,
		HasGameEnd:   true,
	}

	first, err := postGame.FinalizeSession(context.Background(), session)
	if err != nil {
		t.Fatalf("first finalize: %v", err)
	}
	second, err := postGame.FinalizeSession(context.Background(), session)
	if err != nil {
		t.Fatalf("second finalize: %v", err)
	}

	if !first.Persisted {
		t.Fatalf("expected first save to persist")
	}
	if !second.Duplicate {
		t.Fatalf("expected second save to be treated as duplicate")
	}
	if len(persister.saved) != 1 {
		t.Fatalf("expected one persisted match, got %d", len(persister.saved))
	}
}

func TestPostGamePipelineQueuesFailedPersistenceAndRetries(t *testing.T) {
	t.Parallel()

	fail := true
	persister := &stubMatchPersister{
		saveFn: func(_ context.Context, _ PersistedMatch) error {
			if fail {
				return errors.New("storage unavailable")
			}
			return nil
		},
	}
	postGame := newTestPostGamePipeline(t, persister, PostGamePipelineOptions{})

	session := CompletedSession{
		Events: []game.Event{
			game.NewGameStartEvent("game-retry", state.GameModeStandard),
			game.NewTurnStartEvent(types.PlayerTypePlayer),
			game.NewGameEndEvent(types.PlayerTypePlayer),
		},
		Reason:       CompletionReasonGameEnd,
		Complete:     true,
		HasGameStart: true,
		HasGameEnd:   true,
	}

	result, err := postGame.FinalizeSession(context.Background(), session)
	if err != nil {
		t.Fatalf("finalize session: %v", err)
	}
	if !result.Queued || result.PersistError == nil {
		t.Fatalf("expected failed persistence to be queued")
	}

	fail = false
	retryResults, err := postGame.RetryPending(context.Background())
	if err != nil {
		t.Fatalf("retry pending: %v", err)
	}
	if len(retryResults) != 1 || !retryResults[0].Persisted {
		t.Fatalf("expected retry to persist queued match, got %#v", retryResults)
	}
}

func TestPostGamePipelineIntegratesDetectorBuilderAndPersister(t *testing.T) {
	t.Parallel()

	persister := &stubMatchPersister{}
	postGame := newTestPostGamePipeline(t, persister, PostGamePipelineOptions{})
	detector := NewSessionDetector()

	events := []game.Event{
		game.NewGameStartEvent("game-live", state.GameModeStandard),
		game.NewTurnStartEvent(types.PlayerTypePlayer),
		game.NewGameEndEvent(types.PlayerTypeOpponent),
	}

	for _, event := range events {
		for _, session := range detector.Process(event) {
			if _, err := postGame.FinalizeSession(context.Background(), session); err != nil {
				t.Fatalf("finalize session: %v", err)
			}
		}
	}

	if len(persister.saved) != 1 {
		t.Fatalf("expected one persisted match, got %d", len(persister.saved))
	}
	saved := persister.saved[0]
	if saved.State == nil {
		t.Fatalf("expected persisted state")
	}
	if saved.State.GameID != "game-live" {
		t.Fatalf("unexpected game id: %s", saved.State.GameID)
	}
	if !saved.State.IsGameOver() {
		t.Fatalf("expected saved state to be marked as finished")
	}
}

func TestPostGamePipelineFallbackKeyWorksWithoutGameID(t *testing.T) {
	t.Parallel()

	persister := &stubMatchPersister{}
	postGame := newTestPostGamePipeline(t, persister, PostGamePipelineOptions{})

	session := CompletedSession{
		Events: []game.Event{
			testMonitorEvent{eventType: types.EventTypeGameStart, timestamp: time.Unix(10, 0)},
			testMonitorEvent{eventType: types.EventTypeTurnStart, timestamp: time.Unix(11, 0)},
			testMonitorEvent{eventType: types.EventTypeGameEnd, timestamp: time.Unix(12, 0)},
		},
		Reason:       CompletionReasonGameEnd,
		Complete:     true,
		HasGameStart: true,
		HasGameEnd:   true,
	}

	result, err := postGame.FinalizeSession(context.Background(), session)
	if err != nil {
		t.Fatalf("finalize session: %v", err)
	}
	if result.Match.Key == "" {
		t.Fatalf("expected fallback persistence key")
	}
}
