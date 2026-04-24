package game

import (
	"errors"
	"fmt"
	"time"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

// Event описывает любое игровое событие, которое может быть применено к состоянию игры.
type Event interface {
	Type() types.EventType
	Timestamp() time.Time
	Apply(gameState *state.State) error
}

// BaseEvent содержит общие поля для всех событий.
type BaseEvent struct {
	eventType types.EventType
	timestamp time.Time
}

// NewBaseEvent создает базовое событие с текущим временем.
func NewBaseEvent(eventType types.EventType) BaseEvent {
	return BaseEvent{
		eventType: eventType,
		timestamp: time.Now(),
	}
}

// Type возвращает тип события.
func (e BaseEvent) Type() types.EventType {
	return e.eventType
}

// Timestamp возвращает время события.
func (e BaseEvent) Timestamp() time.Time {
	return e.timestamp
}

// GameStartEvent обозначает начало игры.
type GameStartEvent struct {
	BaseEvent
	GameID   string
	GameMode state.GameMode
}

func NewGameStartEvent(gameID string, gameMode state.GameMode) *GameStartEvent {
	return &GameStartEvent{
		BaseEvent: NewBaseEvent(types.EventTypeGameStart),
		GameID:    gameID,
		GameMode:  gameMode,
	}
}

func (e *GameStartEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}

	if e.GameID != "" {
		gameState.GameID = e.GameID
	}
	if e.GameMode != "" {
		gameState.GameMode = e.GameMode
	}
	gameState.StartGame()
	return nil
}

// GameEndEvent обозначает завершение игры.
type GameEndEvent struct {
	BaseEvent
	Winner types.PlayerType
}

func NewGameEndEvent(winner types.PlayerType) *GameEndEvent {
	return &GameEndEvent{
		BaseEvent: NewBaseEvent(types.EventTypeGameEnd),
		Winner:    winner,
	}
}

func (e *GameEndEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}

	gameState.EndGame(e.Winner)
	return nil
}

// TurnStartEvent обозначает начало хода.
type TurnStartEvent struct {
	BaseEvent
	Player types.PlayerType
}

func NewTurnStartEvent(playerType types.PlayerType) *TurnStartEvent {
	return &TurnStartEvent{
		BaseEvent: NewBaseEvent(types.EventTypeTurnStart),
		Player:    playerType,
	}
}

func (e *TurnStartEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}

	gameState.StartTurn(e.Player)
	return nil
}

// TurnEndEvent обозначает завершение хода.
type TurnEndEvent struct {
	BaseEvent
}

func NewTurnEndEvent() *TurnEndEvent {
	return &TurnEndEvent{
		BaseEvent: NewBaseEvent(types.EventTypeTurnEnd),
	}
}

func (e *TurnEndEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}

	gameState.EndTurn()
	return nil
}

// DeckListEvent обозначает определение исходной колоды игрока.
type DeckListEvent struct {
	BaseEvent
	Player types.PlayerType
	Cards  []*card.Card
	Name   string
	DeckID string
	Code   string
}

func NewDeckListEvent(playerType types.PlayerType, cards []*card.Card, name, deckID, code string) *DeckListEvent {
	return &DeckListEvent{
		BaseEvent: NewBaseEvent(types.EventTypeDeckList),
		Player:    playerType,
		Cards:     cards,
		Name:      name,
		DeckID:    deckID,
		Code:      code,
	}
}

func (e *DeckListEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}

	player := gameState.GetPlayer(e.Player)
	if player == nil {
		return fmt.Errorf("player not found: %s", e.Player)
	}

	player.DeckList = append(player.DeckList[:0], e.Cards...)
	if e.Name != "" {
		player.Name = e.Name
	}

	return nil
}

// CardPlayedEvent обозначает розыгрыш карты.
type CardPlayedEvent struct {
	BaseEvent
	Player   types.PlayerType
	Card     *card.Card
	Position int
}

func NewCardPlayedEvent(playerType types.PlayerType, playedCard *card.Card, position int) *CardPlayedEvent {
	return &CardPlayedEvent{
		BaseEvent: NewBaseEvent(types.EventTypeCardPlayed),
		Player:    playerType,
		Card:      playedCard,
		Position:  position,
	}
}

func (e *CardPlayedEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}
	if e.Card == nil {
		return errors.New("card is nil")
	}

	player := gameState.GetPlayer(e.Player)
	if player == nil {
		return fmt.Errorf("player not found: %s", e.Player)
	}

	// Если карта уже была восстановлена в руке, используем обычный переход между зонами.
	if player.Hand.GetCard(e.Card.EntityID) != nil {
		e.Card.Zone = card.ZonePlay
		if !player.PlayCard(e.Card, e.Position) {
			return fmt.Errorf("failed to play card with entity id %d", e.Card.EntityID)
		}
	} else if player.Board.GetCard(e.Card.EntityID) == nil {
		// В реальных логах часть карт/сущностей появляется без полного восстановления руки.
		// В таком случае регистрируем сам факт появления на доске и продолжаем анализ.
		e.Card.Zone = card.ZonePlay
		player.Board.AddCardAt(e.Card, e.Position)
	}

	gameState.TrackCardEntity(e.Card)
	return nil
}

// CardDrawnEvent обозначает взятие карты.
type CardDrawnEvent struct {
	BaseEvent
	Player types.PlayerType
	Card   *card.Card
}

func NewCardDrawnEvent(playerType types.PlayerType, drawnCard *card.Card) *CardDrawnEvent {
	return &CardDrawnEvent{
		BaseEvent: NewBaseEvent(types.EventTypeCardDrawn),
		Player:    playerType,
		Card:      drawnCard,
	}
}

func (e *CardDrawnEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}
	if e.Card == nil {
		return errors.New("card is nil")
	}

	player := gameState.GetPlayer(e.Player)
	if player == nil {
		return fmt.Errorf("player not found: %s", e.Player)
	}

	player.Deck.AddCard(e.Card)
	drawnCard := player.DrawCard()
	if drawnCard == nil {
		return fmt.Errorf("failed to draw card with entity id %d", e.Card.EntityID)
	}
	return nil
}

// AttackEvent обозначает факт атаки.
type AttackEvent struct {
	BaseEvent
	AttackerEntityID int
	TargetEntityID   int
}

func NewAttackEvent(attackerEntityID, targetEntityID int) *AttackEvent {
	return &AttackEvent{
		BaseEvent:        NewBaseEvent(types.EventTypeAttack),
		AttackerEntityID: attackerEntityID,
		TargetEntityID:   targetEntityID,
	}
}

func (e *AttackEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}
	return nil
}

// DamageEvent обозначает нанесение урона.
type DamageEvent struct {
	BaseEvent
	TargetEntityID int
	Amount         int
}

func NewDamageEvent(targetEntityID, amount int) *DamageEvent {
	return &DamageEvent{
		BaseEvent:      NewBaseEvent(types.EventTypeDamage),
		TargetEntityID: targetEntityID,
		Amount:         amount,
	}
}

func (e *DamageEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}
	target := gameState.GetEntity(e.TargetEntityID)
	if target == nil {
		return nil
	}
	target.TakeDamage(e.Amount)
	return nil
}

// HealEvent обозначает исцеление.
type HealEvent struct {
	BaseEvent
	TargetEntityID int
	Amount         int
}

func NewHealEvent(targetEntityID, amount int) *HealEvent {
	return &HealEvent{
		BaseEvent:      NewBaseEvent(types.EventTypeHeal),
		TargetEntityID: targetEntityID,
		Amount:         amount,
	}
}

func (e *HealEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}
	target := gameState.GetEntity(e.TargetEntityID)
	if target == nil {
		return nil
	}
	target.Heal(e.Amount)
	return nil
}

// DeathEvent обозначает смерть сущности.
type DeathEvent struct {
	BaseEvent
	EntityID int
}

func NewDeathEvent(entityID int) *DeathEvent {
	return &DeathEvent{
		BaseEvent: NewBaseEvent(types.EventTypeDeath),
		EntityID:  entityID,
	}
}

func (e *DeathEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}
	target := gameState.GetEntity(e.EntityID)
	if target == nil {
		return nil
	}
	target.Health = 0
	if gameState.Player != nil {
		_ = gameState.Player.RemoveMinion(e.EntityID)
	}
	if gameState.Opponent != nil {
		_ = gameState.Opponent.RemoveMinion(e.EntityID)
	}
	gameState.RemoveTrackedEntity(e.EntityID)
	return nil
}

// EffectTriggeredEvent обозначает срабатывание эффекта.
type EffectTriggeredEvent struct {
	BaseEvent
	SourceEntityID int
	EffectName     string
}

func NewEffectTriggeredEvent(sourceEntityID int, effectName string) *EffectTriggeredEvent {
	return &EffectTriggeredEvent{
		BaseEvent:      NewBaseEvent(types.EventTypeEffectTriggered),
		SourceEntityID: sourceEntityID,
		EffectName:     effectName,
	}
}

func (e *EffectTriggeredEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}
	return nil
}

// DiscoverEvent обозначает событие механики Discover.
type DiscoverEvent struct {
	BaseEvent
	Player       types.PlayerType
	Options      []string
	ChosenCard   string
	ChosenCardID int
}

func NewDiscoverEvent(playerType types.PlayerType, options []string, chosenCard string, chosenCardID int) *DiscoverEvent {
	return &DiscoverEvent{
		BaseEvent:    NewBaseEvent(types.EventTypeDiscover),
		Player:       playerType,
		Options:      options,
		ChosenCard:   chosenCard,
		ChosenCardID: chosenCardID,
	}
}

func (e *DiscoverEvent) Apply(gameState *state.State) error {
	if gameState == nil {
		return errors.New("game state is nil")
	}
	// В текущем доменном слое сохраняем сам факт Discover события.
	return nil
}
