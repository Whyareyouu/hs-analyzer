package analyzer

import (
	"context"
	"fmt"
	"slices"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
)

// EventHandler обрабатывает событие в рамках pipeline.
type EventHandler interface {
	Name() string
	Priority() int
	Handle(ctx context.Context, gameState *state.State, event game.Event) error
}

// Pipeline последовательно применяет обработчики к событиям.
type Pipeline struct {
	handlers []EventHandler
}

// NewPipeline создает пустой pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{
		handlers: make([]EventHandler, 0),
	}
}

// RegisterHandler регистрирует обработчик и поддерживает порядок по приоритету.
func (p *Pipeline) RegisterHandler(handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("handler is nil")
	}

	p.handlers = append(p.handlers, handler)
	slices.SortStableFunc(p.handlers, func(left, right EventHandler) int {
		switch {
		case left.Priority() > right.Priority():
			return -1
		case left.Priority() < right.Priority():
			return 1
		default:
			return 0
		}
	})

	return nil
}

// Handlers возвращает копию зарегистрированных обработчиков.
func (p *Pipeline) Handlers() []EventHandler {
	return append([]EventHandler(nil), p.handlers...)
}

// Process обрабатывает одно событие через все зарегистрированные обработчики.
func (p *Pipeline) Process(ctx context.Context, gameState *state.State, event game.Event) error {
	if gameState == nil {
		return fmt.Errorf("game state is nil")
	}
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	for _, handler := range p.handlers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := handler.Handle(ctx, gameState, event); err != nil {
			return fmt.Errorf("handler %q: %w", handler.Name(), err)
		}
	}

	gameState.AddEvent(event)
	return nil
}

// ProcessAll обрабатывает набор событий по порядку.
func (p *Pipeline) ProcessAll(ctx context.Context, gameState *state.State, events []game.Event) error {
	if gameState == nil {
		return fmt.Errorf("game state is nil")
	}

	for _, event := range events {
		if err := p.Process(ctx, gameState, event); err != nil {
			return err
		}
	}

	return nil
}

// ApplyEventHandler применяет событие к GameState через event.Apply.
type ApplyEventHandler struct{}

// NewApplyEventHandler создает стандартный обработчик применения событий к состоянию.
func NewApplyEventHandler() ApplyEventHandler {
	return ApplyEventHandler{}
}

func (h ApplyEventHandler) Name() string {
	return "apply_event"
}

func (h ApplyEventHandler) Priority() int {
	return 0
}

func (h ApplyEventHandler) Handle(_ context.Context, gameState *state.State, event game.Event) error {
	return event.Apply(gameState)
}
