package state

import (
	"time"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/entity"
	"hsanalyzer/internal/player"
	"hsanalyzer/internal/turn"
	"hsanalyzer/internal/types"
)

// GameMode определяет режим игры
type GameMode string

const (
	GameModeStandard      GameMode = "STANDARD"
	GameModeWild          GameMode = "WILD"
	GameModeArena         GameMode = "ARENA"
	GameModeBattlegrounds GameMode = "BATTLEGROUNDS"
	GameModeTavernBrawl   GameMode = "TAVERN_BRAWL"
	GameModeDuels         GameMode = "DUELS"
	GameModeMercenaries   GameMode = "MERCENARIES"
	GameModeUnknown       GameMode = "UNKNOWN"
)

// State представляет полное состояние игры
type State struct {
	// GameID уникальный идентификатор игры
	GameID string

	// GameMode режим игры
	GameMode GameMode

	// CurrentTurn текущий номер хода
	CurrentTurn int

	// CurrentPlayer текущий игрок
	CurrentPlayer types.PlayerType

	// Player игрок
	Player *player.Player

	// Opponent оппонент
	Opponent *player.Player

	// Turns список ходов
	Turns []*turn.Turn

	// Events список всех событий
	Events []interface{} // Event интерфейс будет определен позже

	// StartTime время начала игры
	StartTime time.Time

	// EndTime время окончания игры
	EndTime *time.Time

	// Winner победитель игры
	Winner *types.PlayerType

	// metadata дополнительные метаданные
	metadata map[string]interface{}

	// entities кэш сущностей по EntityID для игровых объектов кроме героев.
	entities map[int]*entity.Entity
}

// New создает новое состояние игры
func New(gameID string, gameMode GameMode) *State {
	return &State{
		GameID:        gameID,
		GameMode:      gameMode,
		CurrentTurn:   0,
		CurrentPlayer: types.PlayerTypeUnknown,
		Player:        player.New(1, types.PlayerTypePlayer),
		Opponent:      player.New(2, types.PlayerTypeOpponent),
		Turns:         make([]*turn.Turn, 0),
		Events:        make([]interface{}, 0),
		StartTime:     time.Now(),
		metadata:      make(map[string]interface{}),
		entities:      make(map[int]*entity.Entity),
	}
}

// StartGame начинает игру
func (s *State) StartGame() {
	s.StartTime = time.Now()
	s.CurrentTurn = 0
}

// EndGame завершает игру
func (s *State) EndGame(winner types.PlayerType) {
	now := time.Now()
	s.EndTime = &now
	s.Winner = &winner
}

// StartTurn начинает новый ход
func (s *State) StartTurn(p types.PlayerType) *turn.Turn {
	s.CurrentTurn++
	s.CurrentPlayer = p

	t := turn.New(s.CurrentTurn, p)
	s.Turns = append(s.Turns, t)

	// Восстанавливаем ману текущему игроку
	currentPlayer := s.GetPlayer(p)
	if currentPlayer != nil {
		currentPlayer.RefillMana()

		// Увеличиваем максимальную ману (до 10)
		if currentPlayer.MaxMana < 10 {
			currentPlayer.GainMaxMana(1)
			currentPlayer.RefillMana()
		}

		// Сбрасываем атаки миньонов
		s.ResetMinionsAttacks(p)
	}

	return t
}

// EndTurn завершает текущий ход
func (s *State) EndTurn() {
	if len(s.Turns) > 0 {
		currentTurn := s.Turns[len(s.Turns)-1]
		currentTurn.End()
	}
}

// GetCurrentTurn возвращает текущий ход
func (s *State) GetCurrentTurn() *turn.Turn {
	if len(s.Turns) == 0 {
		return nil
	}
	return s.Turns[len(s.Turns)-1]
}

// GetPlayer возвращает игрока по типу
func (s *State) GetPlayer(playerType types.PlayerType) *player.Player {
	if playerType == types.PlayerTypePlayer {
		return s.Player
	}
	if playerType == types.PlayerTypeOpponent {
		return s.Opponent
	}
	return nil
}

// GetOpponent возвращает оппонента для указанного игрока
func (s *State) GetOpponent(playerType types.PlayerType) *player.Player {
	if playerType == types.PlayerTypePlayer {
		return s.Opponent
	}
	if playerType == types.PlayerTypeOpponent {
		return s.Player
	}
	return nil
}

// GetEntity возвращает сущность по EntityID (герои и сущности на доске).
func (s *State) GetEntity(entityID int) *entity.Entity {
	if s.Player != nil && s.Player.Hero != nil && s.Player.Hero.ID == entityID {
		return s.Player.Hero
	}
	if s.Opponent != nil && s.Opponent.Hero != nil && s.Opponent.Hero.ID == entityID {
		return s.Opponent.Hero
	}
	if tracked, ok := s.entities[entityID]; ok {
		return tracked
	}

	if c := s.GetCard(entityID); c != nil {
		tracked := s.TrackCardEntity(c)
		return tracked
	}
	return nil
}

// GetCard возвращает карту по EntityID из зон обоих игроков.
func (s *State) GetCard(entityID int) *card.Card {
	zones := make([][]*card.Card, 0, 8)
	if s.Player != nil {
		zones = append(zones,
			s.Player.Hand.GetCards(),
			s.Player.Deck.GetCards(),
			s.Player.Board.GetCards(),
			s.Player.Graveyard.GetCards(),
		)
	}
	if s.Opponent != nil {
		zones = append(zones,
			s.Opponent.Hand.GetCards(),
			s.Opponent.Deck.GetCards(),
			s.Opponent.Board.GetCards(),
			s.Opponent.Graveyard.GetCards(),
		)
	}

	for _, zoneCards := range zones {
		for _, c := range zoneCards {
			if c != nil && c.EntityID == entityID {
				return c
			}
		}
	}

	return nil
}

// AddEvent добавляет событие в состояние
func (s *State) AddEvent(event interface{}) {
	s.Events = append(s.Events, event)

	// Добавляем событие в текущий ход
	currentTurn := s.GetCurrentTurn()
	if currentTurn != nil {
		currentTurn.AddEvent(event)
	}
}

// TrackCardEntity регистрирует сущность на основе карты и возвращает ее.
func (s *State) TrackCardEntity(c *card.Card) *entity.Entity {
	if s == nil || c == nil {
		return nil
	}

	if tracked, ok := s.entities[c.EntityID]; ok {
		return tracked
	}

	tracked := entity.New(c.EntityID, c.ID, c.Owner)
	tracked.SetStats(0, 1)
	s.entities[c.EntityID] = tracked
	return tracked
}

// RemoveTrackedEntity удаляет сущность из внутреннего кэша.
func (s *State) RemoveTrackedEntity(entityID int) {
	if s == nil {
		return
	}
	delete(s.entities, entityID)
}

// ResetMinionsAttacks сбрасывает атаки миньонов игрока
func (s *State) ResetMinionsAttacks(playerType types.PlayerType) {
	p := s.GetPlayer(playerType)
	if p == nil {
		return
	}

	// В реальной реализации здесь будет сброс атак всех миньонов на доске
	// Пока оставляем заглушку
}

// IsGameOver проверяет, закончена ли игра
func (s *State) IsGameOver() bool {
	return s.EndTime != nil
}

// GetWinner возвращает победителя
func (s *State) GetWinner() *types.PlayerType {
	return s.Winner
}

// SetMetadata устанавливает метаданные
func (s *State) SetMetadata(key string, value interface{}) {
	s.metadata[key] = value
}

// GetMetadata возвращает метаданные
func (s *State) GetMetadata(key string) interface{} {
	return s.metadata[key]
}

// GetTurnCount возвращает количество ходов
func (s *State) GetTurnCount() int {
	return len(s.Turns)
}

// GetEventCount возвращает количество событий
func (s *State) GetEventCount() int {
	return len(s.Events)
}

// GetDuration возвращает длительность игры
func (s *State) GetDuration() time.Duration {
	if s.EndTime != nil {
		return s.EndTime.Sub(s.StartTime)
	}
	return time.Since(s.StartTime)
}
