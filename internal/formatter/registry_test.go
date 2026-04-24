package formatter

import (
	"strings"
	"testing"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
)

type stubFormatter struct {
	name string
}

func (f stubFormatter) Name() string {
	return f.name
}

func (f stubFormatter) Format(*state.State) (string, error) {
	return "", nil
}

func (f stubFormatter) FormatEvent(game.Event) (string, error) {
	return "", nil
}

func TestRegistryRegisterAndGet(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.Register(stubFormatter{name: "stub"}); err != nil {
		t.Fatalf("register formatter: %v", err)
	}

	got, ok := registry.Get("stub")
	if !ok {
		t.Fatalf("expected formatter to be registered")
	}
	if got.Name() != "stub" {
		t.Fatalf("unexpected formatter name: %s", got.Name())
	}
}

func TestRegistryRejectsDuplicate(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.Register(stubFormatter{name: "text"}); err != nil {
		t.Fatalf("register formatter: %v", err)
	}

	err := registry.Register(stubFormatter{name: "text"})
	if err == nil {
		t.Fatalf("expected duplicate formatter error")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistryCanRegisterTextFormatter(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.Register(NewTextFormatter(TextFormatterConfig{})); err != nil {
		t.Fatalf("register text formatter: %v", err)
	}

	got, ok := registry.Get("text")
	if !ok {
		t.Fatalf("expected text formatter")
	}
	if got.Name() != "text" {
		t.Fatalf("unexpected formatter name: %s", got.Name())
	}
}

func TestRegistryCanRegisterJSONFormatter(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.Register(NewJSONFormatter()); err != nil {
		t.Fatalf("register json formatter: %v", err)
	}

	got, ok := registry.Get("json")
	if !ok {
		t.Fatalf("expected json formatter")
	}
	if got.Name() != "json" {
		t.Fatalf("unexpected formatter name: %s", got.Name())
	}
}
