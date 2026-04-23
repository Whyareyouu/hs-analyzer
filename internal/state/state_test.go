package state

import (
	"testing"
	"time"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/entity"
	"hsanalyzer/internal/types"
)

func TestNew(t *testing.T) {
	state := New("game123", GameModeStandard)

	if state.GameID != "game123" {
		t.Errorf("Expected GameID game123, got %s", state.GameID)
	}
	if state.GameMode != GameModeStandard {
		t.Errorf("Expected state.GameMode STANDARD, got %s", state.GameMode)
	}
	if state.CurrentTurn != 0 {
		t.Errorf("Expected CurrentTurn 0, got %d", state.CurrentTurn)
	}
	if state.Player == nil {
		t.Error("Player should not be nil")
	}
	if state.Opponent == nil {
		t.Error("Opponent should not be nil")
	}
	if len(state.Turns) != 0 {
		t.Errorf("Expected 0 turns, got %d", len(state.Turns))
	}
	if len(state.Events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(state.Events))
	}
}

func TestState_StartGame(t *testing.T) {
	state := New("game123", GameModeStandard)

	before := time.Now()
	state.StartGame()
	after := time.Now()

	if state.StartTime.Before(before) || state.StartTime.After(after) {
		t.Error("StartTime should be set to current time")
	}
	if state.CurrentTurn != 0 {
		t.Errorf("Expected CurrentTurn 0, got %d", state.CurrentTurn)
	}
}

func TestState_EndGame(t *testing.T) {
	state := New("game123", GameModeStandard)

	state.EndGame(types.PlayerTypePlayer)

	if state.EndTime == nil {
		t.Error("EndTime should be set")
	}
	if state.Winner == nil {
		t.Error("Winner should be set")
	}
	if *state.Winner != types.PlayerTypePlayer {
		t.Errorf("Expected winner PLAYER, got %s", *state.Winner)
	}
}

func TestState_StartTurn(t *testing.T) {
	state := New("game123", GameModeStandard)
	state.Player.MaxMana = 3
	state.Player.Mana = 0

	turn := state.StartTurn(types.PlayerTypePlayer)

	if turn == nil {
		t.Error("Turn should not be nil")
	}
	if state.CurrentTurn != 1 {
		t.Errorf("Expected CurrentTurn 1, got %d", state.CurrentTurn)
	}
	if state.CurrentPlayer != types.PlayerTypePlayer {
		t.Errorf("Expected CurrentPlayer PLAYER, got %s", state.CurrentPlayer)
	}
	if len(state.Turns) != 1 {
		t.Errorf("Expected 1 turn, got %d", len(state.Turns))
	}
	if state.Player.Mana != 4 {
		t.Errorf("Expected Player Mana 4 (3 + 1 gained), got %d", state.Player.Mana)
	}
	if state.Player.MaxMana != 4 {
		t.Errorf("Expected Player MaxMana 4, got %d", state.Player.MaxMana)
	}
}

func TestState_StartTurn_ManaCap(t *testing.T) {
	state := New("game123", GameModeStandard)
	state.Player.MaxMana = 10
	state.Player.Mana = 0

	state.StartTurn(types.PlayerTypePlayer)

	if state.Player.MaxMana != 10 {
		t.Errorf("Expected Player MaxMana 10 (capped), got %d", state.Player.MaxMana)
	}
	if state.Player.Mana != 10 {
		t.Errorf("Expected Player Mana 10, got %d", state.Player.Mana)
	}
}

func TestState_EndTurn(t *testing.T) {
	state := New("game123", GameModeStandard)
	turn := state.StartTurn(types.PlayerTypePlayer)

	state.EndTurn()

	if turn.EndTime == nil {
		t.Error("Turn EndTime should be set")
	}
}

func TestState_GetCurrentTurn(t *testing.T) {
	state := New("game123", GameModeStandard)

	// Нет ходов
	turn := state.GetCurrentTurn()
	if turn != nil {
		t.Error("Should return nil when no turns")
	}

	// Есть ход
	state.StartTurn(types.PlayerTypePlayer)
	turn = state.GetCurrentTurn()
	if turn == nil {
		t.Error("Should return current turn")
	}
	if turn.Number != 1 {
		t.Errorf("Expected turn number 1, got %d", turn.Number)
	}
}

func TestState_GetPlayer(t *testing.T) {
	state := New("game123", GameModeStandard)

	player := state.GetPlayer(types.PlayerTypePlayer)
	if player == nil {
		t.Error("Should return player")
	}
	if player.Type != types.PlayerTypePlayer {
		t.Errorf("Expected player type PLAYER, got %s", player.Type)
	}

	opponent := state.GetPlayer(types.PlayerTypeOpponent)
	if opponent == nil {
		t.Error("Should return opponent")
	}
	if opponent.Type != types.PlayerTypeOpponent {
		t.Errorf("Expected player type OPPONENT, got %s", opponent.Type)
	}

	unknown := state.GetPlayer(types.PlayerTypeUnknown)
	if unknown != nil {
		t.Error("Should return nil for unknown player type")
	}
}

func TestState_GetOpponent(t *testing.T) {
	state := New("game123", GameModeStandard)

	opponent := state.GetOpponent(types.PlayerTypePlayer)
	if opponent == nil {
		t.Error("Should return opponent")
	}
	if opponent.Type != types.PlayerTypeOpponent {
		t.Errorf("Expected opponent type OPPONENT, got %s", opponent.Type)
	}

	player := state.GetOpponent(types.PlayerTypeOpponent)
	if player == nil {
		t.Error("Should return player")
	}
	if player.Type != types.PlayerTypePlayer {
		t.Errorf("Expected player type PLAYER, got %s", player.Type)
	}
}

func TestState_AddEvent(t *testing.T) {
	state := New("game123", GameModeStandard)
	state.StartTurn(types.PlayerTypePlayer)

	event := "test_event"
	state.AddEvent(event)

	if len(state.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(state.Events))
	}

	turn := state.GetCurrentTurn()
	if len(turn.Events) != 1 {
		t.Errorf("Expected 1 event in turn, got %d", len(turn.Events))
	}
}

func TestState_GetEntity(t *testing.T) {
	state := New("game123", GameModeStandard)

	// Устанавливаем героев
	state.Player.Hero = entity.New(1, "HERO_01", types.PlayerTypePlayer)
	state.Opponent.Hero = entity.New(2, "HERO_02", types.PlayerTypeOpponent)

	// Получаем героя игрока
	entity := state.GetEntity(1)
	if entity == nil {
		t.Error("Should return player hero")
	}
	if entity.ID != 1 {
		t.Errorf("Expected entity ID 1, got %d", entity.ID)
	}

	// Получаем героя оппонента
	entity = state.GetEntity(2)
	if entity == nil {
		t.Error("Should return opponent hero")
	}
	if entity.ID != 2 {
		t.Errorf("Expected entity ID 2, got %d", entity.ID)
	}

	// Несуществующая сущность
	entity = state.GetEntity(999)
	if entity != nil {
		t.Error("Should return nil for non-existent entity")
	}
}

func TestState_GetCard(t *testing.T) {
	state := New("game123", GameModeStandard)

	// Добавляем карту в руку игрока
	card := card.New("CS2_189", 42, types.PlayerTypePlayer)
	state.Player.Hand.AddCard(card)

	// Получаем карту
	found := state.GetCard(42)
	if found == nil {
		t.Error("Should find card")
	}
	if found.EntityID != 42 {
		t.Errorf("Expected EntityID 42, got %d", found.EntityID)
	}

	// Несуществующая карта
	notFound := state.GetCard(999)
	if notFound != nil {
		t.Error("Should return nil for non-existent card")
	}
}

func TestState_IsGameOver(t *testing.T) {
	state := New("game123", GameModeStandard)

	if state.IsGameOver() {
		t.Error("Game should not be over")
	}

	state.EndGame(types.PlayerTypePlayer)

	if !state.IsGameOver() {
		t.Error("Game should be over")
	}
}

func TestState_GetWinner(t *testing.T) {
	state := New("game123", GameModeStandard)

	winner := state.GetWinner()
	if winner != nil {
		t.Error("Winner should be nil")
	}

	state.EndGame(types.PlayerTypePlayer)

	winner = state.GetWinner()
	if winner == nil {
		t.Error("Winner should not be nil")
	}
	if *winner != types.PlayerTypePlayer {
		t.Errorf("Expected winner PLAYER, got %s", *winner)
	}
}

func TestState_SetGetMetadata(t *testing.T) {
	state := New("game123", GameModeStandard)

	state.SetMetadata("test_key", "test_value")

	value := state.GetMetadata("test_key")
	if value == nil {
		t.Error("Metadata should not be nil")
	}
	if value.(string) != "test_value" {
		t.Errorf("Expected test_value, got %v", value)
	}
}

func TestState_GetTurnCount(t *testing.T) {
	state := New("game123", GameModeStandard)

	if state.GetTurnCount() != 0 {
		t.Errorf("Expected 0 turns, got %d", state.GetTurnCount())
	}

	state.StartTurn(types.PlayerTypePlayer)
	state.StartTurn(types.PlayerTypeOpponent)

	if state.GetTurnCount() != 2 {
		t.Errorf("Expected 2 turns, got %d", state.GetTurnCount())
	}
}

func TestState_GetEventCount(t *testing.T) {
	state := New("game123", GameModeStandard)

	if state.GetEventCount() != 0 {
		t.Errorf("Expected 0 events, got %d", state.GetEventCount())
	}

	state.AddEvent("event1")
	state.AddEvent("event2")

	if state.GetEventCount() != 2 {
		t.Errorf("Expected 2 events, got %d", state.GetEventCount())
	}
}

func TestState_GetDuration(t *testing.T) {
	state := New("game123", GameModeStandard)
	state.StartGame()

	time.Sleep(10 * time.Millisecond)

	duration := state.GetDuration()
	if duration < 10*time.Millisecond {
		t.Error("Duration should be at least 10ms")
	}

	state.EndGame(types.PlayerTypePlayer)
	duration = state.GetDuration()
	if duration < 10*time.Millisecond {
		t.Error("Duration should be at least 10ms")
	}
}
