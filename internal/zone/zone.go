package zone

import (
	"hsanalyzer/internal/card"
	"hsanalyzer/internal/types"
)

// Zone представляет игровую зону (рука, доска, колода и т.д.)
type Zone struct {
	// Type - тип зоны
	Type card.ZoneType

	// Cards - карты в зоне
	Cards []*card.Card

	// Owner - владелец зоны
	Owner types.PlayerType
}

// New создает новую зону
func New(zoneType card.ZoneType, owner types.PlayerType) *Zone {
	return &Zone{
		Type:  zoneType,
		Cards: make([]*card.Card, 0),
		Owner: owner,
	}
}

// AddCard добавляет карту в зону
func (z *Zone) AddCard(c *card.Card) {
	c.Zone = z.Type
	c.Position = len(z.Cards)
	z.Cards = append(z.Cards, c)
}

// AddCardAt добавляет карту в зону на определенную позицию
func (z *Zone) AddCardAt(c *card.Card, position int) {
	if position < 0 || position > len(z.Cards) {
		position = len(z.Cards)
	}

	c.Zone = z.Type
	c.Position = position

	// Вставляем карту на нужную позицию
	z.Cards = append(z.Cards[:position], append([]*card.Card{c}, z.Cards[position:]...)...)

	// Обновляем позиции всех карт после вставленной
	z.updatePositions()
}

// RemoveCard удаляет карту из зоны
func (z *Zone) RemoveCard(c *card.Card) bool {
	for i, card := range z.Cards {
		if card.EntityID == c.EntityID {
			z.Cards = append(z.Cards[:i], z.Cards[i+1:]...)
			z.updatePositions()
			return true
		}
	}
	return false
}

// RemoveCardByEntityID удаляет карту из зоны по EntityID
func (z *Zone) RemoveCardByEntityID(entityID int) *card.Card {
	for i, c := range z.Cards {
		if c.EntityID == entityID {
			card := z.Cards[i]
			z.Cards = append(z.Cards[:i], z.Cards[i+1:]...)
			z.updatePositions()
			return card
		}
	}
	return nil
}

// GetCard возвращает карту по EntityID
func (z *Zone) GetCard(entityID int) *card.Card {
	for _, c := range z.Cards {
		if c.EntityID == entityID {
			return c
		}
	}
	return nil
}

// GetCardAt возвращает карту на определенной позиции
func (z *Zone) GetCardAt(position int) *card.Card {
	if position < 0 || position >= len(z.Cards) {
		return nil
	}
	return z.Cards[position]
}

// Count возвращает количество карт в зоне
func (z *Zone) Count() int {
	return len(z.Cards)
}

// IsEmpty проверяет, пуста ли зона
func (z *Zone) IsEmpty() bool {
	return len(z.Cards) == 0
}

// IsFull проверяет, заполнена ли зона (для доски максимум 7 миньонов)
func (z *Zone) IsFull() bool {
	if z.Type == card.ZonePlay {
		return len(z.Cards) >= 7
	}
	if z.Type == card.ZoneHand {
		return len(z.Cards) >= 10
	}
	return false
}

// Clear очищает зону
func (z *Zone) Clear() {
	z.Cards = make([]*card.Card, 0)
}

// GetCards возвращает копию списка карт
func (z *Zone) GetCards() []*card.Card {
	cards := make([]*card.Card, len(z.Cards))
	copy(cards, z.Cards)
	return cards
}

// updatePositions обновляет позиции всех карт в зоне
func (z *Zone) updatePositions() {
	for i, card := range z.Cards {
		card.Position = i
	}
}

// MoveCard перемещает карту на новую позицию внутри зоны
func (z *Zone) MoveCard(entityID int, newPosition int) bool {
	// Находим карту
	var cardIndex int = -1
	for i, c := range z.Cards {
		if c.EntityID == entityID {
			cardIndex = i
			break
		}
	}

	if cardIndex == -1 {
		return false
	}

	if newPosition < 0 || newPosition >= len(z.Cards) {
		return false
	}

	// Если позиция не изменилась
	if cardIndex == newPosition {
		return true
	}

	// Извлекаем карту
	cardToMove := z.Cards[cardIndex]
	z.Cards = append(z.Cards[:cardIndex], z.Cards[cardIndex+1:]...)

	// Вставляем на новую позицию
	z.Cards = append(z.Cards[:newPosition], append([]*card.Card{cardToMove}, z.Cards[newPosition:]...)...)

	// Обновляем позиции
	z.updatePositions()

	return true
}
