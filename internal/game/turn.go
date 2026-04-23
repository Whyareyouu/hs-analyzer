package game

import "time"

// Turn представляет игровой ход
type Turn struct {
	// Number номер хода
	Number int

	// Player игрок, чей это ход
	Player PlayerType

	// Events события в этом ходу
	Events []interface{} // Event интерфейс будет определен позже

	// StartTime время начала хода
	StartTime time.Time

	// EndTime время окончания хода
	EndTime *time.Time

	// CardsPlayed карты, сыгранные в этот ход
	CardsPlayed []*Card

	// AttacksMade атаки, совершенные в этот ход
	AttacksMade []interface{} // Attack будет определен позже

	// CardsDrawn карты, взятые в этот ход
	CardsDrawn []*Card
}

// NewTurn создает новый ход
func NewTurn(number int, player PlayerType) *Turn {
	return &Turn{
		Number:      number,
		Player:      player,
		Events:      make([]interface{}, 0),
		StartTime:   time.Now(),
		CardsPlayed: make([]*Card, 0),
		AttacksMade: make([]interface{}, 0),
		CardsDrawn:  make([]*Card, 0),
	}
}

// AddEvent добавляет событие в ход
func (t *Turn) AddEvent(event interface{}) {
	t.Events = append(t.Events, event)
}

// AddCardPlayed добавляет сыгранную карту
func (t *Turn) AddCardPlayed(card *Card) {
	t.CardsPlayed = append(t.CardsPlayed, card)
}

// AddCardDrawn добавляет взятую карту
func (t *Turn) AddCardDrawn(card *Card) {
	t.CardsDrawn = append(t.CardsDrawn, card)
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
