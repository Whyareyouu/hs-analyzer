package zone

import (
	"testing"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/types"
)

func TestNew(t *testing.T) {
	zone := New(card.ZoneHand, types.PlayerTypePlayer)

	if zone.Type != card.ZoneHand {
		t.Errorf("Expected Type HAND, got %s", zone.Type)
	}
	if zone.Owner != types.PlayerTypePlayer {
		t.Errorf("Expected Owner PLAYER, got %s", zone.Owner)
	}
	if len(zone.Cards) != 0 {
		t.Errorf("Expected 0 cards, got %d", len(zone.Cards))
	}
}

func TestZone_AddCard(t *testing.T) {
	zone := New(card.ZoneHand, types.PlayerTypePlayer)
	c := card.New("CS2_189", 42, types.PlayerTypePlayer)

	zone.AddCard(c)

	if zone.Count() != 1 {
		t.Errorf("Expected 1 card, got %d", zone.Count())
	}
	if c.Zone != card.ZoneHand {
		t.Errorf("Expected card zone HAND, got %s", c.Zone)
	}
	if c.Position != 0 {
		t.Errorf("Expected card position 0, got %d", c.Position)
	}
}

func TestZone_AddCardAt(t *testing.T) {
	zone := New(card.ZonePlay, types.PlayerTypePlayer)
	card1 := card.New("CARD_1", 1, types.PlayerTypePlayer)
	card2 := card.New("CARD_2", 2, types.PlayerTypePlayer)
	card3 := card.New("CARD_3", 3, types.PlayerTypePlayer)

	zone.AddCard(card1)
	zone.AddCard(card2)
	zone.AddCardAt(card3, 1)

	if zone.Count() != 3 {
		t.Errorf("Expected 3 cards, got %d", zone.Count())
	}
	if zone.Cards[0].EntityID != 1 {
		t.Errorf("Expected card 1 at position 0")
	}
	if zone.Cards[1].EntityID != 3 {
		t.Errorf("Expected card 3 at position 1")
	}
	if zone.Cards[2].EntityID != 2 {
		t.Errorf("Expected card 2 at position 2")
	}
}

func TestZone_RemoveCard(t *testing.T) {
	zone := New(card.ZoneHand, types.PlayerTypePlayer)
	card1 := card.New("CARD_1", 1, types.PlayerTypePlayer)
	card2 := card.New("CARD_2", 2, types.PlayerTypePlayer)

	zone.AddCard(card1)
	zone.AddCard(card2)

	removed := zone.RemoveCard(card1)

	if !removed {
		t.Error("Card should be removed")
	}
	if zone.Count() != 1 {
		t.Errorf("Expected 1 card, got %d", zone.Count())
	}
	if zone.Cards[0].EntityID != 2 {
		t.Error("Wrong card remaining")
	}
}

func TestZone_RemoveCardByEntityID(t *testing.T) {
	zone := New(card.ZoneHand, types.PlayerTypePlayer)
	card1 := card.New("CARD_1", 1, types.PlayerTypePlayer)
	card2 := card.New("CARD_2", 2, types.PlayerTypePlayer)

	zone.AddCard(card1)
	zone.AddCard(card2)

	removed := zone.RemoveCardByEntityID(1)

	if removed == nil {
		t.Error("Card should be removed")
	}
	if removed.EntityID != 1 {
		t.Errorf("Expected removed card EntityID 1, got %d", removed.EntityID)
	}
	if zone.Count() != 1 {
		t.Errorf("Expected 1 card, got %d", zone.Count())
	}
}

func TestZone_GetCard(t *testing.T) {
	zone := New(card.ZoneHand, types.PlayerTypePlayer)
	c := card.New("CS2_189", 42, types.PlayerTypePlayer)

	zone.AddCard(c)

	found := zone.GetCard(42)
	if found == nil {
		t.Error("Card should be found")
	}
	if found.EntityID != 42 {
		t.Errorf("Expected EntityID 42, got %d", found.EntityID)
	}

	notFound := zone.GetCard(999)
	if notFound != nil {
		t.Error("Card should not be found")
	}
}

func TestZone_GetCardAt(t *testing.T) {
	zone := New(card.ZoneHand, types.PlayerTypePlayer)
	card1 := card.New("CARD_1", 1, types.PlayerTypePlayer)
	card2 := card.New("CARD_2", 2, types.PlayerTypePlayer)

	zone.AddCard(card1)
	zone.AddCard(card2)

	found := zone.GetCardAt(1)
	if found == nil {
		t.Error("Card should be found")
	}
	if found.EntityID != 2 {
		t.Errorf("Expected EntityID 2, got %d", found.EntityID)
	}

	notFound := zone.GetCardAt(10)
	if notFound != nil {
		t.Error("Card should not be found")
	}
}

func TestZone_Count(t *testing.T) {
	zone := New(card.ZoneHand, types.PlayerTypePlayer)

	if zone.Count() != 0 {
		t.Errorf("Expected 0 cards, got %d", zone.Count())
	}

	zone.AddCard(card.New("CARD_1", 1, types.PlayerTypePlayer))
	zone.AddCard(card.New("CARD_2", 2, types.PlayerTypePlayer))

	if zone.Count() != 2 {
		t.Errorf("Expected 2 cards, got %d", zone.Count())
	}
}

func TestZone_IsEmpty(t *testing.T) {
	zone := New(card.ZoneHand, types.PlayerTypePlayer)

	if !zone.IsEmpty() {
		t.Error("Zone should be empty")
	}

	zone.AddCard(card.New("CARD_1", 1, types.PlayerTypePlayer))

	if zone.IsEmpty() {
		t.Error("Zone should not be empty")
	}
}

func TestZone_IsFull(t *testing.T) {
	// Test board (max 7 minions)
	board := New(card.ZonePlay, types.PlayerTypePlayer)

	if board.IsFull() {
		t.Error("Board should not be full")
	}

	for i := 0; i < 7; i++ {
		board.AddCard(card.New("CARD", i, types.PlayerTypePlayer))
	}

	if !board.IsFull() {
		t.Error("Board should be full")
	}

	// Test hand (max 10 cards)
	hand := New(card.ZoneHand, types.PlayerTypePlayer)

	if hand.IsFull() {
		t.Error("Hand should not be full")
	}

	for i := 0; i < 10; i++ {
		hand.AddCard(card.New("CARD", i, types.PlayerTypePlayer))
	}

	if !hand.IsFull() {
		t.Error("Hand should be full")
	}
}

func TestZone_Clear(t *testing.T) {
	zone := New(card.ZoneHand, types.PlayerTypePlayer)
	zone.AddCard(card.New("CARD_1", 1, types.PlayerTypePlayer))
	zone.AddCard(card.New("CARD_2", 2, types.PlayerTypePlayer))

	zone.Clear()

	if !zone.IsEmpty() {
		t.Error("Zone should be empty after clear")
	}
}

func TestZone_GetCards(t *testing.T) {
	zone := New(card.ZoneHand, types.PlayerTypePlayer)
	card1 := card.New("CARD_1", 1, types.PlayerTypePlayer)
	card2 := card.New("CARD_2", 2, types.PlayerTypePlayer)

	zone.AddCard(card1)
	zone.AddCard(card2)

	cards := zone.GetCards()

	if len(cards) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(cards))
	}

	// Проверяем, что это копия
	cards[0] = nil
	if zone.Cards[0] == nil {
		t.Error("Original zone cards should not be affected")
	}
}

func TestZone_MoveCard(t *testing.T) {
	zone := New(card.ZonePlay, types.PlayerTypePlayer)
	card1 := card.New("CARD_1", 1, types.PlayerTypePlayer)
	card2 := card.New("CARD_2", 2, types.PlayerTypePlayer)
	card3 := card.New("CARD_3", 3, types.PlayerTypePlayer)

	zone.AddCard(card1)
	zone.AddCard(card2)
	zone.AddCard(card3)

	// Перемещаем карту 1 на позицию 2
	moved := zone.MoveCard(1, 2)

	if !moved {
		t.Error("Card should be moved")
	}
	if zone.Cards[0].EntityID != 2 {
		t.Error("Card 2 should be at position 0")
	}
	if zone.Cards[1].EntityID != 3 {
		t.Error("Card 3 should be at position 1")
	}
	if zone.Cards[2].EntityID != 1 {
		t.Error("Card 1 should be at position 2")
	}
}

func TestZone_MoveCard_InvalidEntityID(t *testing.T) {
	zone := New(card.ZonePlay, types.PlayerTypePlayer)
	zone.AddCard(card.New("CARD_1", 1, types.PlayerTypePlayer))

	moved := zone.MoveCard(999, 0)

	if moved {
		t.Error("Should not move non-existent card")
	}
}

func TestZone_MoveCard_InvalidPosition(t *testing.T) {
	zone := New(card.ZonePlay, types.PlayerTypePlayer)
	zone.AddCard(card.New("CARD_1", 1, types.PlayerTypePlayer))

	moved := zone.MoveCard(1, 10)

	if moved {
		t.Error("Should not move to invalid position")
	}
}
