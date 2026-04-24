package parser

import (
	"context"
	"io"

	"hsanalyzer/internal/game"
)

// LogType определяет тип лога, который умеет обрабатывать парсер.
type LogType string

const (
	LogTypePower         LogType = "Power"
	LogTypeZone          LogType = "Zone"
	LogTypeLoadingScreen LogType = "LoadingScreen"
	LogTypeDeck          LogType = "Deck"
)

// Parser определяет интерфейс для парсинга логов Hearthstone в игровые события.
type Parser interface {
	Parse(ctx context.Context, reader io.Reader) ([]game.Event, error)
	Name() string
	LogType() LogType
}
