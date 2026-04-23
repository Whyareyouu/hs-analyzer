package player

import (
	"hsanalyzer/internal/card"
	"hsanalyzer/internal/entity"
	"hsanalyzer/internal/types"
	"hsanalyzer/internal/zone"
)

// Player представляет игрока
type Player struct {
	// ID игрока
	ID int

	// Name имя игрока
	Name string

	// Type тип игрока (игрок или оппонент)
	Type types.PlayerType

	// Hero герой игрока
	Hero *entity.Entity

	// HeroPower сила героя
	HeroPower *card.Card

	// Weapon оружие героя
	Weapon *card.Card

	// Hand рука игрока
	Hand *zone.Zone

	// Deck колода игрока
	Deck *zone.Zone

	// Board доска игрока
	Board *zone.Zone

	// Graveyard кладбище игрока
	Graveyard *zone.Zone

	// Mana текущая мана
	Mana int

	// MaxMana максимальная мана
	MaxMana int

	// OverloadMana перегрузка маны
	OverloadMana int

	// DeckList изначальный список колоды
	DeckList []*card.Card
}

// New создает нового игрока
func New(id int, playerType types.PlayerType) *Player {
	return &Player{
		ID:        id,
		Type:      playerType,
		Hand:      zone.New(card.ZoneHand, playerType),
		Deck:      zone.New(card.ZoneDeck, playerType),
		Board:     zone.New(card.ZonePlay, playerType),
		Graveyard: zone.New(card.ZoneGraveyard, playerType),
		DeckList:  make([]*card.Card, 0),
		Mana:      0,
		MaxMana:   0,
	}
}

// DrawCard берет карту из колоды в руку
func (p *Player) DrawCard() *card.Card {
	if p.Deck.IsEmpty() {
		return nil
	}

	// Берем первую карту из колоды
	c := p.Deck.GetCardAt(0)
	if c == nil {
		return nil
	}

	p.Deck.RemoveCard(c)

	// Если рука полная, карта сгорает
	if p.Hand.IsFull() {
		p.Graveyard.AddCard(c)
		return c
	}

	p.Hand.AddCard(c)
	return c
}

// PlayCard разыгрывает карту из руки
func (p *Player) PlayCard(c *card.Card, position int) bool {
	if !p.Hand.RemoveCard(c) {
		return false
	}

	// Если это миньон, добавляем на доску
	if c.Zone == card.ZonePlay {
		if p.Board.IsFull() {
			return false
		}
		p.Board.AddCardAt(c, position)
	} else {
		// Заклинания и другие карты идут на кладбище
		p.Graveyard.AddCard(c)
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
func (p *Player) GetMinion(entityID int) *entity.Entity {
	c := p.Board.GetCard(entityID)
	if c == nil {
		return nil
	}
	// В реальной реализации здесь будет связь с Entity
	// Пока возвращаем nil, так как нужна связь с GameState
	return nil
}

// RemoveMinion удаляет миньона с доски
func (p *Player) RemoveMinion(entityID int) *card.Card {
	c := p.Board.RemoveCardByEntityID(entityID)
	if c != nil {
		p.Graveyard.AddCard(c)
	}
	return c
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
