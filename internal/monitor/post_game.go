package monitor

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
)

// EventProcessor applies a list of events to a game state.
type EventProcessor interface {
	ProcessAll(ctx context.Context, gameState *state.State, events []game.Event) error
}

// MatchPersister persists one finalized match into the canonical storage layer.
type MatchPersister interface {
	SaveMatch(ctx context.Context, match PersistedMatch) error
}

// PersistedMatch is the canonical handoff object from live monitoring into storage.
type PersistedMatch struct {
	Key      string
	State    *state.State
	Events   []game.Event
	Complete bool
	Reason   CompletionReason
}

// FinalizationResult describes what happened to one completed session.
type FinalizationResult struct {
	Match        PersistedMatch
	Persisted    bool
	Queued       bool
	Duplicate    bool
	Skipped      bool
	PersistError error
}

// PostGamePipelineOptions controls how incomplete sessions are handled.
type PostGamePipelineOptions struct {
	PersistIncompleteFlush bool
}

// StateBuilder reconstructs one state from a session event stream.
type StateBuilder struct {
	processor EventProcessor
}

// NewStateBuilder creates a builder that reuses the existing analyzer pipeline semantics.
func NewStateBuilder(processor EventProcessor) (*StateBuilder, error) {
	if processor == nil {
		return nil, errors.New("event processor is nil")
	}
	return &StateBuilder{processor: processor}, nil
}

// Build reconstructs a state from the given events.
func (b *StateBuilder) Build(ctx context.Context, events []game.Event) (*state.State, error) {
	if b == nil {
		return nil, errors.New("state builder is nil")
	}
	if len(events) == 0 {
		return nil, errors.New("events are empty")
	}

	gameState := state.New("", state.GameModeUnknown)
	if err := b.processor.ProcessAll(ctx, gameState, events); err != nil {
		return nil, fmt.Errorf("process session events: %w", err)
	}

	return gameState, nil
}

// PostGamePipeline takes completed sessions, reconstructs matches and persists them.
type PostGamePipeline struct {
	builder   *StateBuilder
	persister MatchPersister
	options   PostGamePipelineOptions

	persistedKeys map[string]struct{}
	pending       map[string]PersistedMatch
}

// NewPostGamePipeline creates a post-game finalization pipeline.
func NewPostGamePipeline(builder *StateBuilder, persister MatchPersister, options PostGamePipelineOptions) (*PostGamePipeline, error) {
	if builder == nil {
		return nil, errors.New("state builder is nil")
	}
	if persister == nil {
		return nil, errors.New("match persister is nil")
	}

	return &PostGamePipeline{
		builder:       builder,
		persister:     persister,
		options:       options,
		persistedKeys: make(map[string]struct{}),
		pending:       make(map[string]PersistedMatch),
	}, nil
}

// FinalizeSession builds and persists a completed session if it is sufficiently complete.
func (p *PostGamePipeline) FinalizeSession(ctx context.Context, session CompletedSession) (FinalizationResult, error) {
	if p == nil {
		return FinalizationResult{}, errors.New("post-game pipeline is nil")
	}
	if len(session.Events) == 0 {
		return FinalizationResult{}, errors.New("session events are empty")
	}

	if !p.shouldPersist(session) {
		return FinalizationResult{
			Match: PersistedMatch{
				Events:   cloneEvents(session.Events),
				Complete: session.Complete,
				Reason:   session.Reason,
			},
			Skipped: true,
		}, nil
	}

	gameState, err := p.builder.Build(ctx, session.Events)
	if err != nil {
		return FinalizationResult{}, err
	}

	match := PersistedMatch{
		State:    gameState,
		Events:   cloneEvents(session.Events),
		Complete: session.Complete,
		Reason:   session.Reason,
	}
	match.Key = buildPersistedMatchKey(match)

	if _, ok := p.persistedKeys[match.Key]; ok {
		return FinalizationResult{Match: match, Duplicate: true}, nil
	}
	if _, ok := p.pending[match.Key]; ok {
		return FinalizationResult{Match: match, Duplicate: true, Queued: true}, nil
	}

	if err := p.persister.SaveMatch(ctx, match); err != nil {
		p.pending[match.Key] = match
		return FinalizationResult{
			Match:        match,
			Queued:       true,
			PersistError: err,
		}, nil
	}

	p.persistedKeys[match.Key] = struct{}{}
	delete(p.pending, match.Key)

	return FinalizationResult{
		Match:     match,
		Persisted: true,
	}, nil
}

// RetryPending retries persistence for matches that previously failed to save.
func (p *PostGamePipeline) RetryPending(ctx context.Context) ([]FinalizationResult, error) {
	if p == nil {
		return nil, errors.New("post-game pipeline is nil")
	}

	results := make([]FinalizationResult, 0, len(p.pending))
	for key, match := range p.pending {
		if _, ok := p.persistedKeys[key]; ok {
			delete(p.pending, key)
			results = append(results, FinalizationResult{
				Match:     match,
				Persisted: true,
				Duplicate: true,
			})
			continue
		}

		if err := p.persister.SaveMatch(ctx, match); err != nil {
			results = append(results, FinalizationResult{
				Match:        match,
				Queued:       true,
				PersistError: err,
			})
			continue
		}

		p.persistedKeys[key] = struct{}{}
		delete(p.pending, key)
		results = append(results, FinalizationResult{
			Match:     match,
			Persisted: true,
		})
	}

	return results, nil
}

func (p *PostGamePipeline) shouldPersist(session CompletedSession) bool {
	if !session.HasGameStart {
		return false
	}
	if session.Complete {
		return true
	}

	switch session.Reason {
	case CompletionReasonNextGameStart:
		return true
	case CompletionReasonFlush, CompletionReasonStreamEnd:
		return p.options.PersistIncompleteFlush
	default:
		return false
	}
}

func buildPersistedMatchKey(match PersistedMatch) string {
	if match.State != nil && match.State.GameID != "" {
		return "game:" + match.State.GameID
	}

	hasher := fnv.New64a()
	for _, event := range match.Events {
		if event == nil {
			continue
		}
		_, _ = hasher.Write([]byte(event.Type().String()))
		_, _ = hasher.Write([]byte(event.Timestamp().UTC().Format("2006-01-02T15:04:05.000000000Z07:00")))
	}
	_, _ = hasher.Write([]byte(strconv.Itoa(len(match.Events))))

	return "events:" + strconv.FormatUint(hasher.Sum64(), 16)
}
