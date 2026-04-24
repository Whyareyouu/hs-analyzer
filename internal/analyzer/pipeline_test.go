package analyzer

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

type stubHandler struct {
	name     string
	priority int
	handleFn func(context.Context, *state.State, game.Event) error
}

func (h stubHandler) Name() string {
	return h.name
}

func (h stubHandler) Priority() int {
	return h.priority
}

func (h stubHandler) Handle(ctx context.Context, gameState *state.State, event game.Event) error {
	if h.handleFn != nil {
		return h.handleFn(ctx, gameState, event)
	}
	return nil
}

func TestNewPipeline(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline()
	if pipeline == nil {
		t.Fatalf("expected pipeline")
	}
	if len(pipeline.Handlers()) != 0 {
		t.Fatalf("unexpected handlers count: %d", len(pipeline.Handlers()))
	}
}

func TestPipelineRegisterHandlerSortsByPriority(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline()
	handlers := []EventHandler{
		stubHandler{name: "low", priority: 10},
		stubHandler{name: "high", priority: 100},
		stubHandler{name: "mid", priority: 50},
	}

	for _, handler := range handlers {
		if err := pipeline.RegisterHandler(handler); err != nil {
			t.Fatalf("register handler: %v", err)
		}
	}

	got := pipeline.Handlers()
	names := []string{got[0].Name(), got[1].Name(), got[2].Name()}
	want := []string{"high", "mid", "low"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("unexpected handler order: got=%v want=%v", names, want)
	}
}

func TestPipelineProcessExecutesHandlersAndAddsEvent(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline()
	gameState := state.New("game-1", state.GameModeUnknown)
	callOrder := make([]string, 0, 2)

	if err := pipeline.RegisterHandler(stubHandler{
		name:     "audit",
		priority: 100,
		handleFn: func(_ context.Context, _ *state.State, _ game.Event) error {
			callOrder = append(callOrder, "audit")
			return nil
		},
	}); err != nil {
		t.Fatalf("register audit handler: %v", err)
	}

	if err := pipeline.RegisterHandler(stubHandler{
		name:     "apply",
		priority: 0,
		handleFn: func(_ context.Context, gameState *state.State, event game.Event) error {
			callOrder = append(callOrder, "apply")
			return event.Apply(gameState)
		},
	}); err != nil {
		t.Fatalf("register apply handler: %v", err)
	}

	event := game.NewGameStartEvent("game-1", state.GameModeStandard)
	if err := pipeline.Process(context.Background(), gameState, event); err != nil {
		t.Fatalf("process event: %v", err)
	}

	if !reflect.DeepEqual(callOrder, []string{"audit", "apply"}) {
		t.Fatalf("unexpected call order: %v", callOrder)
	}
	if gameState.GameMode != state.GameModeStandard {
		t.Fatalf("unexpected game mode: %s", gameState.GameMode)
	}
	if gameState.GetEventCount() != 1 {
		t.Fatalf("unexpected event count: %d", gameState.GetEventCount())
	}
}

func TestPipelineProcessStopsOnHandlerError(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline()
	gameState := state.New("game-1", state.GameModeUnknown)

	if err := pipeline.RegisterHandler(stubHandler{
		name:     "broken",
		priority: 100,
		handleFn: func(_ context.Context, _ *state.State, _ game.Event) error {
			return errors.New("boom")
		},
	}); err != nil {
		t.Fatalf("register broken handler: %v", err)
	}

	if err := pipeline.RegisterHandler(NewApplyEventHandler()); err != nil {
		t.Fatalf("register apply handler: %v", err)
	}

	err := pipeline.Process(context.Background(), gameState, game.NewTurnStartEvent(types.PlayerTypePlayer))
	if err == nil {
		t.Fatalf("expected handler error")
	}
	if !strings.Contains(err.Error(), `handler "broken": boom`) {
		t.Fatalf("unexpected error: %v", err)
	}
	if gameState.GetEventCount() != 0 {
		t.Fatalf("event should not be added after handler error")
	}
}

func TestPipelineProcessAllAppliesEventsInOrder(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline()
	if err := pipeline.RegisterHandler(NewApplyEventHandler()); err != nil {
		t.Fatalf("register apply handler: %v", err)
	}

	gameState := state.New("game-1", state.GameModeUnknown)
	events := []game.Event{
		game.NewGameStartEvent("game-1", state.GameModeStandard),
		game.NewTurnStartEvent(types.PlayerTypePlayer),
		game.NewTurnEndEvent(),
	}

	if err := pipeline.ProcessAll(context.Background(), gameState, events); err != nil {
		t.Fatalf("process all: %v", err)
	}

	if gameState.GameMode != state.GameModeStandard {
		t.Fatalf("unexpected game mode: %s", gameState.GameMode)
	}
	if gameState.GetTurnCount() != 1 {
		t.Fatalf("unexpected turn count: %d", gameState.GetTurnCount())
	}
	if gameState.GetEventCount() != 3 {
		t.Fatalf("unexpected event count: %d", gameState.GetEventCount())
	}
	currentTurn := gameState.GetCurrentTurn()
	if currentTurn == nil || currentTurn.EndTime == nil {
		t.Fatalf("expected turn to be ended")
	}
}

func TestPipelineRejectsNilInputs(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline()

	if err := pipeline.RegisterHandler(nil); err == nil {
		t.Fatalf("expected nil handler error")
	}
	if err := pipeline.Process(context.Background(), nil, game.NewTurnEndEvent()); err == nil {
		t.Fatalf("expected nil state error")
	}
	if err := pipeline.Process(context.Background(), state.New("game-1", state.GameModeUnknown), nil); err == nil {
		t.Fatalf("expected nil event error")
	}
}
