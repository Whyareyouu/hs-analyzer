package game

import "time"

// GameState представляет полное состояние игры
type GameState struct {
	// GameID уникальный идентификатор игры
	GameID string

	// GameMode режим игры
	GameMode GameMode

	// CurrentTurn текущий номер хода
	CurrentTurn int

	// CurrentPlayer текущий игрок
	CurrentPlayer PlayerType

	// Player игрок
	Player *Player

	// Opponent оппонент
	Opponent *Player

	// Turns список ходов
	Turns []*Turn

	// Events список всех событий
	Events []interface{} // Event интерфейс будет определен позже

	// StartTime время начала игры
	StartTime time.Time

	// EndTime время окончания игры
	EndTime *time.Time

	// Winner победитель игры
	Winner *PlayerType

	// metadata дополнительные метаданные
	metadata map[string]interface{}
}

// NewGameState создает новое состояние игры
func NewGameState(gameID string, gameMode GameMode) *GameState {
	return &GameState{
		GameID:        gameID,
		GameMode:      gameMode,
		CurrentTurn:   0,
		CurrentPlayer: PlayerTypeUnknown,
		Player:        NewPlayer(1, PlayerTypePlayer),
		Opponent:      NewPlayer(2, PlayerTypeOpponent),
		Turns:         make([]*Turn, 0),
		Events:        make([]interface{}, 0),
		StartTime:     time.Now(),
		metadata:      make(map[string]interface{}),
	}
}

// StartGame начинает игру
func (gs *GameState) StartGame() {
	gs.StartTime = time.Now()
	gs.CurrentTurn = 0
}

// EndGame завершает игру
func (gs *GameState) EndGame(winner PlayerType) {
	now := time.Now()
	gs.EndTime = &now
	gs.Winner = &winner
}

// StartTurn начинает новый ход
func (gs *GameState) StartTurn(player PlayerType) *Turn {
	gs.CurrentTurn++
	gs.CurrentPlayer = player

	turn := NewTurn(gs.CurrentTurn, player)
	gs.Turns = append(gs.Turns, turn)

	// Восстанавливаем ману текущему игроку
	currentPlayer := gs.GetPlayer(player)
	if currentPlayer != nil {
		currentPlayer.RefillMana()

		// Увеличиваем максимальную ману (до 10)
		if currentPlayer.MaxMana < 10 {
			currentPlayer.GainMaxMana(1)
			currentPlayer.RefillMana()
		}

		// Сбрасываем атаки миньонов
		gs.ResetMinionsAttacks(player)
	}

	return turn
}

// EndTurn завершает текущий ход
func (gs *GameState) EndTurn() {
	if len(gs.Turns) > 0 {
		currentTurn := gs.Turns[len(gs.Turns)-1]
		currentTurn.End()
	}
}

// GetCurrentTurn возвращает текущий ход
func (gs *GameState) GetCurrentTurn() *Turn {
	if len(gs.Turns) == 0 {
		return nil
	}
	return gs.Turns[len(gs.Turns)-1]
}

// GetPlayer возвращает игрока по типу
func (gs *GameState) GetPlayer(playerType PlayerType) *Player {
	if playerType == PlayerTypePlayer {
		return gs.Player
	}
	if playerType == PlayerTypeOpponent {
		return gs.Opponent
	}
	return nil
}

// GetOpponent возвращает оппонента для указанного игрока
func (gs *GameState) GetOpponent(playerType PlayerType) *Player {
	if playerType == PlayerTypePlayer {
		return gs.Opponent
	}
	if playerType == PlayerTypeOpponent {
		return gs.Player
	}
	return nil
}

// AddEvent добавляет событие в состояние
func (gs *GameState) AddEvent(event interface{}) {
	gs.Events = append(gs.Events, event)

	// Добавляем событие в текущий ход
	currentTurn := gs.GetCurrentTurn()
	if currentTurn != nil {
		currentTurn.AddEvent(event)
	}
}

// GetEntity возвращает сущность по ID
func (gs *GameState) GetEntity(entityID int) *Entity {
	// Проверяем героев
	if gs.Player.Hero != nil && gs.Player.Hero.ID == entityID {
		return gs.Player.Hero
	}
	if gs.Opponent.Hero != nil && gs.Opponent.Hero.ID == entityID {
		return gs.Opponent.Hero
	}

	// Проверяем миньонов на досках
	// Здесь нужна более сложная логика для связи Card и Entity
	// Пока возвращаем nil
	return nil
}

// GetCard возвращает карту по EntityID
func (gs *GameState) GetCard(entityID int) *Card {
	// Проверяем все зоны игрока
	if card := gs.Player.Hand.GetCard(entityID); card != nil {
		return card
	}
	if card := gs.Player.Deck.GetCard(entityID); card != nil {
		return card
	}
	if card := gs.Player.Board.GetCard(entityID); card != nil {
		return card
	}
	if card := gs.Player.Graveyard.GetCard(entityID); card != nil {
		return card
	}

	// Проверяем все зоны оппонента
	if card := gs.Opponent.Hand.GetCard(entityID); card != nil {
		return card
	}
	if card := gs.Opponent.Deck.GetCard(entityID); card != nil {
		return card
	}
	if card := gs.Opponent.Board.GetCard(entityID); card != nil {
		return card
	}
	if card := gs.Opponent.Graveyard.GetCard(entityID); card != nil {
		return card
	}

	return nil
}

// ResetMinionsAttacks сбрасывает атаки миньонов игрока
func (gs *GameState) ResetMinionsAttacks(playerType PlayerType) {
	player := gs.GetPlayer(playerType)
	if player == nil {
		return
	}

	// В реальной реализации здесь будет сброс атак всех миньонов на доске
	// Пока оставляем заглушку
}

// IsGameOver проверяет, закончена ли игра
func (gs *GameState) IsGameOver() bool {
	return gs.EndTime != nil
}

// GetWinner возвращает победителя
func (gs *GameState) GetWinner() *PlayerType {
	return gs.Winner
}

// SetMetadata устанавливает метаданные
func (gs *GameState) SetMetadata(key string, value interface{}) {
	gs.metadata[key] = value
}

// GetMetadata возвращает метаданные
func (gs *GameState) GetMetadata(key string) interface{} {
	return gs.metadata[key]
}

// GetTurnCount возвращает количество ходов
func (gs *GameState) GetTurnCount() int {
	return len(gs.Turns)
}

// GetEventCount возвращает количество событий
func (gs *GameState) GetEventCount() int {
	return len(gs.Events)
}

// GetDuration возвращает длительность игры
func (gs *GameState) GetDuration() time.Duration {
	if gs.EndTime != nil {
		return gs.EndTime.Sub(gs.StartTime)
	}
	return time.Since(gs.StartTime)
}
