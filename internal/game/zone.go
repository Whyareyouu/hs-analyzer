package game

// Zone представляет игровую зону (рука, доска, колода и т.д.)
type Zone struct {
	// Type - тип зоны
	Type ZoneType

	// Cards - карты в зоне
	Cards []*Card

	// Owner - владелец зоны
	Owner PlayerType
}

// NewZone создает новую зону
func NewZone(zoneType ZoneType, owner PlayerType) *Zone {
	return &Zone{
		Type:  zoneType,
		Cards: make([]*Card, 0),
		Owner: owner,
	}
}

// AddCard добавляет карту в зону
func (z *Zone) AddCard(card *Card) {
	card.Zone = z.Type
	card.Position = len(z.Cards)
	z.Cards = append(z.Cards, card)
}

// AddCardAt добавляет карту в зону на определенную позицию
func (z *Zone) AddCardAt(card *Card, position int) {
	if position < 0 || position > len(z.Cards) {
		position = len(z.Cards)
	}

	card.Zone = z.Type
	card.Position = position

	// Вставляем карту на нужную позицию
	z.Cards = append(z.Cards[:position], append([]*Card{card}, z.Cards[position:]...)...)

	// Обновляем позиции всех карт после вставленной
	z.updatePositions()
}

// RemoveCard удаляет карту из зоны
func (z *Zone) RemoveCard(card *Card) bool {
	for i, c := range z.Cards {
		if c.EntityID == card.EntityID {
			z.Cards = append(z.Cards[:i], z.Cards[i+1:]...)
			z.updatePositions()
			return true
		}
	}
	return false
}

// RemoveCardByEntityID удаляет карту из зоны по EntityID
func (z *Zone) RemoveCardByEntityID(entityID int) *Card {
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
func (z *Zone) GetCard(entityID int) *Card {
	for _, c := range z.Cards {
		if c.EntityID == entityID {
			return c
		}
	}
	return nil
}

// GetCardAt возвращает карту на определенной позиции
func (z *Zone) GetCardAt(position int) *Card {
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
	if z.Type == ZonePlay {
		return len(z.Cards) >= 7
	}
	if z.Type == ZoneHand {
		return len(z.Cards) >= 10
	}
	return false
}

// Clear очищает зону
func (z *Zone) Clear() {
	z.Cards = make([]*Card, 0)
}

// GetCards возвращает копию списка карт
func (z *Zone) GetCards() []*Card {
	cards := make([]*Card, len(z.Cards))
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
	card := z.Cards[cardIndex]
	z.Cards = append(z.Cards[:cardIndex], z.Cards[cardIndex+1:]...)

	// Вставляем на новую позицию
	z.Cards = append(z.Cards[:newPosition], append([]*Card{card}, z.Cards[newPosition:]...)...)

	// Обновляем позиции
	z.updatePositions()

	return true
}
