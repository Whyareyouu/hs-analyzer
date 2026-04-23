package turn

import (
	"time"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/types"
)

// Turn представляет игровой ход
type Turn struct {
	// Number номер хода
	Number int

	// Player игрок, чей это ход
	Player types.PlayerType

	// Events события в этом ходу
	Events []interface{} // Event интерфейс будет определен позже

	// StartTime время начала хода
	StartTime time.Time

	// EndTime время окончания хода
	EndTime *time.Time

	// CardsPlayed карты, сыгранные в этот ход
	CardsPlayed []*card.Card

	// AttacksMade атаки, совершенные в этот ход
	AttacksMade []interface{} // Attack будет определен позже

	// CardsDrawn карты, взятые в этот ход
	CardsDrawn []*card.Card
}

// New создает новый ход
func New(number int, player types.PlayerType) *Turn {
	return &Turn{
		Number:      number,
		Player:      player,
		Events:      make([]interface{}, 0),
		StartTime:   time.Now(),
		CardsPlayed: make([]*card.Card, 0),
		AttacksMade: make([]interface{}, 0),
		CardsDrawn:  make([]*card.Card, 0),
	}
}

// AddEvent добавляет событие в ход
func (t *Turn) AddEvent(event interface{}) {
	t.Events = append(t.Events, event)
}

// AddCardPlayed добавляет сыгранную карту
func (t *Turn) AddCardPlayed(c *card.Card) {
	t.CardsPlayed = append(t.CardsPlayed, c)
}

// AddCardDrawn добавляет взятую карту
func (t *Turn) AddCardDrawn(c *card.Card) {
	t.CardsDrawn = append(t.CardsDrawn, c)
}

// AddAttack добавляет атаку
func (t *Turn) AddAttack(attack interface{}) {
	t.AttacksMade = append(t.AttacksMade, attack)
}

// End завершает ход
func (t *Turn) End() {
	now := time.Now()
	t.EndTime = &now
}

// IsEnded проверяет, завершен ли ход
func (t *Turn) IsEnded() bool {
	return t.EndTime != nil
}

// GetDuration возвращает длительность хода
func (t *Turn) GetDuration() time.Duration {
	if t.EndTime != nil {
		return t.EndTime.Sub(t.StartTime)
	}
	return time.Since(t.StartTime)
}

// GetEventCount возвращает количество событий в ходу
func (t *Turn) GetEventCount() int {
	return len(t.Events)
}

// GetCardsPlayedCount возвращает количество сыгранных карт
func (t *Turn) GetCardsPlayedCount() int {
	return len(t.CardsPlayed)
}

// GetCardsDrawnCount возвращает количество взятых карт
func (t *Turn) GetCardsDrawnCount() int {
	return len(t.CardsDrawn)
}

// GetAttacksCount возвращает количество атак
func (t *Turn) GetAttacksCount() int {
	return len(t.AttacksMade)
}
