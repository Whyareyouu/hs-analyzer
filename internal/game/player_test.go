package game

import "testing"

func TestNewPlayer(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)

	if player.ID != 1 {
		t.Errorf("Expected ID 1, got %d", player.ID)
	}
	if player.Type != PlayerTypePlayer {
		t.Errorf("Expected Type PLAYER, got %s", player.Type)
	}
	if player.Hand == nil {
		t.Error("Hand should not be nil")
	}
	if player.Deck == nil {
		t.Error("Deck should not be nil")
	}
	if player.Board == nil {
		t.Error("Board should not be nil")
	}
	if player.Graveyard == nil {
		t.Error("Graveyard should not be nil")
	}
	if player.Mana != 0 {
		t.Errorf("Expected Mana 0, got %d", player.Mana)
	}
	if player.MaxMana != 0 {
		t.Errorf("Expected MaxMana 0, got %d", player.MaxMana)
	}
}

func TestPlayer_DrawCard(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)

	// Добавляем карту в колоду
	card := NewCard("CS2_189", 42, PlayerTypePlayer)
	player.Deck.AddCard(card)

	// Берем карту
	drawnCard := player.DrawCard()

	if drawnCard == nil {
		t.Error("Should draw a card")
	}
	if drawnCard.EntityID != 42 {
		t.Errorf("Expected EntityID 42, got %d", drawnCard.EntityID)
	}
	if player.Hand.Count() != 1 {
		t.Errorf("Expected 1 card in hand, got %d", player.Hand.Count())
	}
	if player.Deck.Count() != 0 {
		t.Errorf("Expected 0 cards in deck, got %d", player.Deck.Count())
	}
}

func TestPlayer_DrawCard_EmptyDeck(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)

	drawnCard := player.DrawCard()

	if drawnCard != nil {
		t.Error("Should not draw a card from empty deck")
	}
}

func TestPlayer_DrawCard_FullHand(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)

	// Заполняем руку
	for i := 0; i < 10; i++ {
		player.Hand.AddCard(NewCard("CARD", i, PlayerTypePlayer))
	}

	// Добавляем карту в колоду
	card := NewCard("CS2_189", 42, PlayerTypePlayer)
	player.Deck.AddCard(card)

	// Пытаемся взять карту
	drawnCard := player.DrawCard()

	if drawnCard == nil {
		t.Error("Should return the burned card")
	}
	if player.Hand.Count() != 10 {
		t.Errorf("Expected 10 cards in hand, got %d", player.Hand.Count())
	}
	if player.Graveyard.Count() != 1 {
		t.Errorf("Expected 1 card in graveyard (burned), got %d", player.Graveyard.Count())
	}
}

func TestPlayer_SpendMana(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)
	player.Mana = 5

	success := player.SpendMana(3)

	if !success {
		t.Error("Should successfully spend mana")
	}
	if player.Mana != 2 {
		t.Errorf("Expected Mana 2, got %d", player.Mana)
	}
}

func TestPlayer_SpendMana_NotEnough(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)
	player.Mana = 2

	success := player.SpendMana(5)

	if success {
		t.Error("Should not spend more mana than available")
	}
	if player.Mana != 2 {
		t.Errorf("Expected Mana 2, got %d", player.Mana)
	}
}

func TestPlayer_RefillMana(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)
	player.MaxMana = 5
	player.Mana = 0

	player.RefillMana()

	if player.Mana != 5 {
		t.Errorf("Expected Mana 5, got %d", player.Mana)
	}
}

func TestPlayer_RefillMana_WithOverload(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)
	player.MaxMana = 5
	player.Mana = 0
	player.OverloadMana = 2

	player.RefillMana()

	if player.Mana != 3 {
		t.Errorf("Expected Mana 3 (5 - 2 overload), got %d", player.Mana)
	}
	if player.OverloadMana != 0 {
		t.Errorf("Expected OverloadMana 0 (cleared), got %d", player.OverloadMana)
	}
}

func TestPlayer_GainMaxMana(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)
	player.MaxMana = 5

	player.GainMaxMana(2)

	if player.MaxMana != 7 {
		t.Errorf("Expected MaxMana 7, got %d", player.MaxMana)
	}
}

func TestPlayer_GainMaxMana_Cap(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)
	player.MaxMana = 9

	player.GainMaxMana(5)

	if player.MaxMana != 10 {
		t.Errorf("Expected MaxMana 10 (capped), got %d", player.MaxMana)
	}
}

func TestPlayer_SetMaxMana(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)

	player.SetMaxMana(7)

	if player.MaxMana != 7 {
		t.Errorf("Expected MaxMana 7, got %d", player.MaxMana)
	}
}

func TestPlayer_AddOverload(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)

	player.AddOverload(2)

	if player.OverloadMana != 2 {
		t.Errorf("Expected OverloadMana 2, got %d", player.OverloadMana)
	}
}

func TestPlayer_RemoveMinion(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)
	card := NewCard("CS2_189", 42, PlayerTypePlayer)
	player.Board.AddCard(card)

	removed := player.RemoveMinion(42)

	if removed == nil {
		t.Error("Should remove minion")
	}
	if player.Board.Count() != 0 {
		t.Errorf("Expected 0 minions on board, got %d", player.Board.Count())
	}
	if player.Graveyard.Count() != 1 {
		t.Errorf("Expected 1 card in graveyard, got %d", player.Graveyard.Count())
	}
}

func TestPlayer_GetHandSize(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)
	player.Hand.AddCard(NewCard("CARD_1", 1, PlayerTypePlayer))
	player.Hand.AddCard(NewCard("CARD_2", 2, PlayerTypePlayer))

	size := player.GetHandSize()

	if size != 2 {
		t.Errorf("Expected hand size 2, got %d", size)
	}
}

func TestPlayer_GetBoardSize(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)
	player.Board.AddCard(NewCard("CARD_1", 1, PlayerTypePlayer))
	player.Board.AddCard(NewCard("CARD_2", 2, PlayerTypePlayer))
	player.Board.AddCard(NewCard("CARD_3", 3, PlayerTypePlayer))

	size := player.GetBoardSize()

	if size != 3 {
		t.Errorf("Expected board size 3, got %d", size)
	}
}

func TestPlayer_GetDeckSize(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)
	for i := 0; i < 30; i++ {
		player.Deck.AddCard(NewCard("CARD", i, PlayerTypePlayer))
	}

	size := player.GetDeckSize()

	if size != 30 {
		t.Errorf("Expected deck size 30, got %d", size)
	}
}

func TestPlayer_IsAlive(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)

	// Без героя
	if player.IsAlive() {
		t.Error("Player without hero should not be alive")
	}

	// С живым героем
	player.Hero = NewEntity(1, "HERO_01", PlayerTypePlayer)
	player.Hero.Health = 30
	if !player.IsAlive() {
		t.Error("Player with alive hero should be alive")
	}

	// С мертвым героем
	player.Hero.Health = 0
	if player.IsAlive() {
		t.Error("Player with dead hero should not be alive")
	}
}

func TestPlayer_IsDead(t *testing.T) {
	player := NewPlayer(1, PlayerTypePlayer)

	// Без героя
	if !player.IsDead() {
		t.Error("Player without hero should be dead")
	}

	// С живым героем
	player.Hero = NewEntity(1, "HERO_01", PlayerTypePlayer)
	player.Hero.Health = 30
	if player.IsDead() {
		t.Error("Player with alive hero should not be dead")
	}

	// С мертвым героем
	player.Hero.Health = 0
	if !player.IsDead() {
		t.Error("Player with dead hero should be dead")
	}
}
