package game

// Player представляет игрока
type Player struct {
	// ID игрока
	ID int

	// Name имя игрока
	Name string

	// Type тип игрока (игрок или оппонент)
	Type PlayerType

	// Hero герой игрока
	Hero *Entity

	// HeroPower сила героя
	HeroPower *Card

	// Weapon оружие героя
	Weapon *Card

	// Hand рука игрока
	Hand *Zone

	// Deck колода игрока
	Deck *Zone

	// Board доска игрока
	Board *Zone

	// Graveyard кладбище игрока
	Graveyard *Zone

	// Mana текущая мана
	Mana int

	// MaxMana максимальная мана
	MaxMana int

	// OverloadMana перегрузка маны
	OverloadMana int

	// DeckList изначальный список колоды
	DeckList []*Card
}

// NewPlayer создает нового игрока
func NewPlayer(id int, playerType PlayerType) *Player {
	return &Player{
		ID:        id,
		Type:      playerType,
		Hand:      NewZone(ZoneHand, playerType),
		Deck:      NewZone(ZoneDeck, playerType),
		Board:     NewZone(ZonePlay, playerType),
		Graveyard: NewZone(ZoneGraveyard, playerType),
		DeckList:  make([]*Card, 0),
		Mana:      0,
		MaxMana:   0,
	}
}

// DrawCard берет карту из колоды в руку
func (p *Player) DrawCard() *Card {
	if p.Deck.IsEmpty() {
		return nil
	}

	// Берем первую карту из колоды
	card := p.Deck.GetCardAt(0)
	if card == nil {
		return nil
	}

	p.Deck.RemoveCard(card)

	// Если рука полная, карта сгорает
	if p.Hand.IsFull() {
		p.Graveyard.AddCard(card)
		return card
	}

	p.Hand.AddCard(card)
	return card
}

// PlayCard разыгрывает карту из руки
func (p *Player) PlayCard(card *Card, position int) bool {
	if !p.Hand.RemoveCard(card) {
		return false
	}

	// Если это миньон, добавляем на доску
	if card.Zone == ZonePlay {
		if p.Board.IsFull() {
			return false
		}
		p.Board.AddCardAt(card, position)
	} else {
		// Заклинания и другие карты идут на кладбище
		p.Graveyard.AddCard(card)
	}

	return true
}

// SpendMana тратит ману
func (p *Player) SpendMana(amount int) bool {
	if p.Mana < amount {
		return false
	}
	p.Mana -= amount
	return true
}

// RefillMana восстанавливает ману в начале хода
func (p *Player) RefillMana() {
	p.Mana = p.MaxMana - p.OverloadMana
	if p.Mana < 0 {
		p.Mana = 0
	}
	p.OverloadMana = 0
}

// GainMaxMana увеличивает максимальную ману
func (p *Player) GainMaxMana(amount int) {
	p.MaxMana += amount
	if p.MaxMana > 10 {
		p.MaxMana = 10
	}
}

// SetMaxMana устанавливает максимальную ману
func (p *Player) SetMaxMana(amount int) {
	p.MaxMana = amount
	if p.MaxMana > 10 {
		p.MaxMana = 10
	}
}

// AddOverload добавляет перегрузку
func (p *Player) AddOverload(amount int) {
	p.OverloadMana += amount
}

// GetMinion возвращает миньона на доске по EntityID
func (p *Player) GetMinion(entityID int) *Entity {
	card := p.Board.GetCard(entityID)
	if card == nil {
		return nil
	}
	// В реальной реализации здесь будет связь с Entity
	// Пока возвращаем nil, так как нужна связь с GameState
	return nil
}

// RemoveMinion удаляет миньона с доски
func (p *Player) RemoveMinion(entityID int) *Card {
	card := p.Board.RemoveCardByEntityID(entityID)
	if card != nil {
		p.Graveyard.AddCard(card)
	}
	return card
}

// GetHandSize возвращает размер руки
func (p *Player) GetHandSize() int {
	return p.Hand.Count()
}

// GetBoardSize возвращает количество миньонов на доске
func (p *Player) GetBoardSize() int {
	return p.Board.Count()
}

// GetDeckSize возвращает размер колоды
func (p *Player) GetDeckSize() int {
	return p.Deck.Count()
}

// IsAlive проверяет, жив ли игрок
func (p *Player) IsAlive() bool {
	return p.Hero != nil && p.Hero.IsAlive()
}

// IsDead проверяет, мертв ли игрок
func (p *Player) IsDead() bool {
	return p.Hero == nil || p.Hero.IsDead()
}
