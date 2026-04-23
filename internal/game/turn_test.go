package game

import (
	"testing"
	"time"
)

func TestNewTurn(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	if turn.Number != 1 {
		t.Errorf("Expected Number 1, got %d", turn.Number)
	}
	if turn.Player != PlayerTypePlayer {
		t.Errorf("Expected Player PLAYER, got %s", turn.Player)
	}
	if len(turn.Events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(turn.Events))
	}
	if len(turn.CardsPlayed) != 0 {
		t.Errorf("Expected 0 cards played, got %d", len(turn.CardsPlayed))
	}
	if len(turn.AttacksMade) != 0 {
		t.Errorf("Expected 0 attacks made, got %d", len(turn.AttacksMade))
	}
	if len(turn.CardsDrawn) != 0 {
		t.Errorf("Expected 0 cards drawn, got %d", len(turn.CardsDrawn))
	}
	if turn.EndTime != nil {
		t.Error("EndTime should be nil")
	}
}

func TestTurn_AddEvent(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	event1 := "event1"
	event2 := "event2"

	turn.AddEvent(event1)
	turn.AddEvent(event2)

	if len(turn.Events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(turn.Events))
	}
	if turn.Events[0].(string) != "event1" {
		t.Errorf("Expected first event to be event1, got %v", turn.Events[0])
	}
	if turn.Events[1].(string) != "event2" {
		t.Errorf("Expected second event to be event2, got %v", turn.Events[1])
	}
}

func TestTurn_AddCardPlayed(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	card1 := NewCard("CARD_1", 1, PlayerTypePlayer)
	card2 := NewCard("CARD_2", 2, PlayerTypePlayer)

	turn.AddCardPlayed(card1)
	turn.AddCardPlayed(card2)

	if len(turn.CardsPlayed) != 2 {
		t.Errorf("Expected 2 cards played, got %d", len(turn.CardsPlayed))
	}
	if turn.CardsPlayed[0].EntityID != 1 {
		t.Errorf("Expected first card EntityID 1, got %d", turn.CardsPlayed[0].EntityID)
	}
	if turn.CardsPlayed[1].EntityID != 2 {
		t.Errorf("Expected second card EntityID 2, got %d", turn.CardsPlayed[1].EntityID)
	}
}

func TestTurn_AddCardDrawn(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	card1 := NewCard("CARD_1", 1, PlayerTypePlayer)
	card2 := NewCard("CARD_2", 2, PlayerTypePlayer)

	turn.AddCardDrawn(card1)
	turn.AddCardDrawn(card2)

	if len(turn.CardsDrawn) != 2 {
		t.Errorf("Expected 2 cards drawn, got %d", len(turn.CardsDrawn))
	}
	if turn.CardsDrawn[0].EntityID != 1 {
		t.Errorf("Expected first card EntityID 1, got %d", turn.CardsDrawn[0].EntityID)
	}
	if turn.CardsDrawn[1].EntityID != 2 {
		t.Errorf("Expected second card EntityID 2, got %d", turn.CardsDrawn[1].EntityID)
	}
}

func TestTurn_AddAttack(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	attack1 := "attack1"
	attack2 := "attack2"

	turn.AddAttack(attack1)
	turn.AddAttack(attack2)

	if len(turn.AttacksMade) != 2 {
		t.Errorf("Expected 2 attacks, got %d", len(turn.AttacksMade))
	}
	if turn.AttacksMade[0].(string) != "attack1" {
		t.Errorf("Expected first attack to be attack1, got %v", turn.AttacksMade[0])
	}
	if turn.AttacksMade[1].(string) != "attack2" {
		t.Errorf("Expected second attack to be attack2, got %v", turn.AttacksMade[1])
	}
}

func TestTurn_End(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	if turn.IsEnded() {
		t.Error("Turn should not be ended initially")
	}

	before := time.Now()
	turn.End()
	after := time.Now()

	if !turn.IsEnded() {
		t.Error("Turn should be ended")
	}
	if turn.EndTime == nil {
		t.Error("EndTime should be set")
	}
	if turn.EndTime.Before(before) || turn.EndTime.After(after) {
		t.Error("EndTime should be set to current time")
	}
}

func TestTurn_IsEnded(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	if turn.IsEnded() {
		t.Error("Turn should not be ended")
	}

	turn.End()

	if !turn.IsEnded() {
		t.Error("Turn should be ended")
	}
}

func TestTurn_GetDuration(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	time.Sleep(10 * time.Millisecond)

	duration := turn.GetDuration()
	if duration < 10*time.Millisecond {
		t.Error("Duration should be at least 10ms")
	}

	turn.End()
	duration = turn.GetDuration()
	if duration < 10*time.Millisecond {
		t.Error("Duration should be at least 10ms")
	}
}

func TestTurn_GetEventCount(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	if turn.GetEventCount() != 0 {
		t.Errorf("Expected 0 events, got %d", turn.GetEventCount())
	}

	turn.AddEvent("event1")
	turn.AddEvent("event2")
	turn.AddEvent("event3")

	if turn.GetEventCount() != 3 {
		t.Errorf("Expected 3 events, got %d", turn.GetEventCount())
	}
}

func TestTurn_GetCardsPlayedCount(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	if turn.GetCardsPlayedCount() != 0 {
		t.Errorf("Expected 0 cards played, got %d", turn.GetCardsPlayedCount())
	}

	turn.AddCardPlayed(NewCard("CARD_1", 1, PlayerTypePlayer))
	turn.AddCardPlayed(NewCard("CARD_2", 2, PlayerTypePlayer))

	if turn.GetCardsPlayedCount() != 2 {
		t.Errorf("Expected 2 cards played, got %d", turn.GetCardsPlayedCount())
	}
}

func TestTurn_GetCardsDrawnCount(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	if turn.GetCardsDrawnCount() != 0 {
		t.Errorf("Expected 0 cards drawn, got %d", turn.GetCardsDrawnCount())
	}

	turn.AddCardDrawn(NewCard("CARD_1", 1, PlayerTypePlayer))
	turn.AddCardDrawn(NewCard("CARD_2", 2, PlayerTypePlayer))
	turn.AddCardDrawn(NewCard("CARD_3", 3, PlayerTypePlayer))

	if turn.GetCardsDrawnCount() != 3 {
		t.Errorf("Expected 3 cards drawn, got %d", turn.GetCardsDrawnCount())
	}
}

func TestTurn_GetAttacksCount(t *testing.T) {
	turn := NewTurn(1, PlayerTypePlayer)

	if turn.GetAttacksCount() != 0 {
		t.Errorf("Expected 0 attacks, got %d", turn.GetAttacksCount())
	}

	turn.AddAttack("attack1")
	turn.AddAttack("attack2")
	turn.AddAttack("attack3")
	turn.AddAttack("attack4")

	if turn.GetAttacksCount() != 4 {
		t.Errorf("Expected 4 attacks, got %d", turn.GetAttacksCount())
	}
}

func TestTurn_CompleteFlow(t *testing.T) {
	// Тест полного потока хода
	turn := NewTurn(5, PlayerTypeOpponent)

	// Добавляем события
	turn.AddEvent("turn_start")
	turn.AddCardDrawn(NewCard("CARD_1", 1, PlayerTypeOpponent))
	turn.AddCardPlayed(NewCard("CARD_2", 2, PlayerTypeOpponent))
	turn.AddCardPlayed(NewCard("CARD_3", 3, PlayerTypeOpponent))
	turn.AddAttack("attack1")
	turn.AddAttack("attack2")
	turn.AddEvent("turn_end")

	// Проверяем состояние
	if turn.Number != 5 {
		t.Errorf("Expected turn number 5, got %d", turn.Number)
	}
	if turn.Player != PlayerTypeOpponent {
		t.Errorf("Expected player OPPONENT, got %s", turn.Player)
	}
	if turn.GetEventCount() != 2 {
		t.Errorf("Expected 2 events, got %d", turn.GetEventCount())
	}
	if turn.GetCardsDrawnCount() != 1 {
		t.Errorf("Expected 1 card drawn, got %d", turn.GetCardsDrawnCount())
	}
	if turn.GetCardsPlayedCount() != 2 {
		t.Errorf("Expected 2 cards played, got %d", turn.GetCardsPlayedCount())
	}
	if turn.GetAttacksCount() != 2 {
		t.Errorf("Expected 2 attacks, got %d", turn.GetAttacksCount())
	}

	// Завершаем ход
	turn.End()

	if !turn.IsEnded() {
		t.Error("Turn should be ended")
	}
}
