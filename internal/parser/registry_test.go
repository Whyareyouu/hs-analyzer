package parser

import (
	"context"
	"io"
	"strings"
	"testing"

	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
)

type stubParser struct {
	name    string
	logType LogType
}

func (p stubParser) Parse(context.Context, io.Reader) ([]game.Event, error) {
	return nil, nil
}

func (p stubParser) Name() string {
	return p.name
}

func (p stubParser) LogType() LogType {
	return p.logType
}

func TestRegistryRegisterAndGet(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	parser := stubParser{name: "stub", logType: LogTypePower}

	if err := registry.Register(parser); err != nil {
		t.Fatalf("register parser: %v", err)
	}

	got, ok := registry.Get(LogTypePower)
	if !ok {
		t.Fatalf("expected parser to be registered")
	}
	if got.Name() != "stub" {
		t.Fatalf("unexpected parser name: %s", got.Name())
	}
}

func TestRegistryRejectsDuplicateLogType(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	first := stubParser{name: "first", logType: LogTypePower}
	second := stubParser{name: "second", logType: LogTypePower}

	if err := registry.Register(first); err != nil {
		t.Fatalf("register first parser: %v", err)
	}

	err := registry.Register(second)
	if err == nil {
		t.Fatalf("expected duplicate registration error")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistryCanRegisterDeckParser(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	parser := NewDeckParser(carddb.NewJSONDatabase())

	if err := registry.Register(parser); err != nil {
		t.Fatalf("register deck parser: %v", err)
	}

	got, ok := registry.Get(LogTypeDeck)
	if !ok {
		t.Fatalf("expected deck parser to be registered")
	}
	if got.Name() != "deck_parser" {
		t.Fatalf("unexpected parser name: %s", got.Name())
	}
}
