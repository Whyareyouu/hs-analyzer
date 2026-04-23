package game

// Card представляет карту в игре
type Card struct {
	// ID карты (например, "CS2_189" для Эльфийского лучника)
	ID string

	// EntityID - уникальный ID сущности в игре
	EntityID int

	// Info - информация о карте из базы данных
	Info interface{} // CardInfo будет определен в carddb пакете

	// Cost - текущая стоимость карты
	Cost int

	// OriginalCost - изначальная стоимость карты
	OriginalCost int

	// Owner - владелец карты
	Owner PlayerType

	// Controller - текущий контроллер карты (может отличаться от владельца)
	Controller PlayerType

	// Zone - текущая зона карты
	Zone ZoneType

	// Position - позиция в зоне (для доски и руки)
	Position int

	// CreatedBy - источник создания карты (если создана эффектом)
	CreatedBy string
}

// NewCard создает новую карту
func NewCard(id string, entityID int, owner PlayerType) *Card {
	return &Card{
		ID:         id,
		EntityID:   entityID,
		Owner:      owner,
		Controller: owner,
		Zone:       ZoneUnknown,
		Position:   -1,
	}
}

// IsInHand проверяет, находится ли карта в руке
func (c *Card) IsInHand() bool {
	return c.Zone == ZoneHand
}

// IsInPlay проверяет, находится ли карта на доске
func (c *Card) IsInPlay() bool {
	return c.Zone == ZonePlay
}

// IsInDeck проверяет, находится ли карта в колоде
func (c *Card) IsInDeck() bool {
	return c.Zone == ZoneDeck
}

// IsInGraveyard проверяет, находится ли карта на кладбище
func (c *Card) IsInGraveyard() bool {
	return c.Zone == ZoneGraveyard
}

// MoveTo перемещает карту в указанную зону
func (c *Card) MoveTo(zone ZoneType, position int) {
	c.Zone = zone
	c.Position = position
}

// SetCost устанавливает стоимость карты
func (c *Card) SetCost(cost int) {
	c.Cost = cost
}

// ResetCost сбрасывает стоимость карты к изначальной
func (c *Card) ResetCost() {
	c.Cost = c.OriginalCost
}
