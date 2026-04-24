package formatter

import (
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
)

// Formatter описывает форматтер результата анализа и отдельных событий.
type Formatter interface {
	Name() string
	Format(gameState *state.State) (string, error)
	FormatEvent(event game.Event) (string, error)
}
