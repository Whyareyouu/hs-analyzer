package game

import "testing"

func TestNewCard(t *testing.T) {
	card := NewCard("CS2_189", 42, PlayerTypePlayer)

	if card.ID != "CS2_189" {
		t.Errorf("Expected ID CS2_189, got %s", card.ID)
	}
	if card.EntityID != 42 {
		t.Errorf("Expected EntityID 42, got %d", card.EntityID)
	}
	if card.Owner != PlayerTypePlayer {
		t.Errorf("Expected Owner PLAYER, got %s", card.Owner)
	}
	if card.Controller != PlayerTypePlayer {
		t.Errorf("Expected Controller PLAYER, got %s", card.Controller)
	}
	if card.Zone != ZoneUnknown {
		t.Errorf("Expected Zone UNKNOWN, got %s", card.Zone)
	}
	if card.Position != -1 {
		t.Errorf("Expected Position -1, got %d", card.Position)
	}
}

func TestCard_IsInHand(t *testing.T) {
	card := NewCard("CS2_189", 42, PlayerTypePlayer)

	if card.IsInHand() {
		t.Error("Card should not be in hand initially")
	}

	card.Zone = ZoneHand
	if !card.IsInHand() {
		t.Error("Card should be in hand")
	}
}

func TestCard_IsInPlay(t *testing.T) {
	card := NewCard("CS2_189", 42, PlayerTypePlayer)

	if card.IsInPlay() {
		t.Error("Card should not be in play initially")
	}

	card.Zone = ZonePlay
	if !card.IsInPlay() {
		t.Error("Card should be in play")
	}
}

func TestCard_IsInDeck(t *testing.T) {
	card := NewCard("CS2_189", 42, PlayerTypePlayer)

	if card.IsInDeck() {
		t.Error("Card should not be in deck initially")
	}

	card.Zone = ZoneDeck
	if !card.IsInDeck() {
		t.Error("Card should be in deck")
	}
}

func TestCard_IsInGraveyard(t *testing.T) {
	card := NewCard("CS2_189", 42, PlayerTypePlayer)

	if card.IsInGraveyard() {
		t.Error("Card should not be in graveyard initially")
	}

	card.Zone = ZoneGraveyard
	if !card.IsInGraveyard() {
		t.Error("Card should be in graveyard")
	}
}

func TestCard_MoveTo(t *testing.T) {
	card := NewCard("CS2_189", 42, PlayerTypePlayer)

	card.MoveTo(ZoneHand, 3)

	if card.Zone != ZoneHand {
		t.Errorf("Expected Zone HAND, got %s", card.Zone)
	}
	if card.Position != 3 {
		t.Errorf("Expected Position 3, got %d", card.Position)
	}
}

func TestCard_SetCost(t *testing.T) {
	card := NewCard("CS2_189", 42, PlayerTypePlayer)
	card.OriginalCost = 5
	card.Cost = 5

	card.SetCost(3)

	if card.Cost != 3 {
		t.Errorf("Expected Cost 3, got %d", card.Cost)
	}
}

func TestCard_ResetCost(t *testing.T) {
	card := NewCard("CS2_189", 42, PlayerTypePlayer)
	card.OriginalCost = 5
	card.Cost = 3

	card.ResetCost()

	if card.Cost != 5 {
		t.Errorf("Expected Cost 5, got %d", card.Cost)
	}
}
