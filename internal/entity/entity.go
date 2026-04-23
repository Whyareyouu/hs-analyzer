package entity

import "hsanalyzer/internal/types"

// Entity представляет сущность на доске (миньон, герой)
type Entity struct {
	// ID - уникальный ID сущности в игре
	ID int

	// CardID - ID карты (например, "CS2_189")
	CardID string

	// Card - информация о карте из базы данных
	Card interface{} // CardInfo будет определен в carddb пакете

	// Attack - текущая атака
	Attack int

	// Health - текущее здоровье
	Health int

	// MaxHealth - максимальное здоровье
	MaxHealth int

	// Position - позиция на доске (0-6)
	Position int

	// Owner - владелец сущности
	Owner types.PlayerType

	// Attributes - атрибуты сущности (Taunt, Divine Shield и т.д.)
	Attributes Attributes

	// Enchantments - активные зачарования на сущности
	Enchantments []*Enchantment

	// CreatedBy - сущность, создавшая эту сущность
	CreatedBy *Entity

	// CreatedByEffect - название эффекта, создавшего сущность
	CreatedByEffect string
}

// Attributes содержит атрибуты сущности
type Attributes struct {
	// Taunt - провокация
	Taunt bool

	// DivineShield - божественный щит
	DivineShield bool

	// Stealth - скрытность
	Stealth bool

	// Windfury - неистовство ветра
	Windfury bool

	// Charge - натиск
	Charge bool

	// Rush - рывок
	Rush bool

	// Lifesteal - похищение здоровья
	Lifesteal bool

	// Poisonous - ядовитость
	Poisonous bool

	// Frozen - заморожен
	Frozen bool

	// Silenced - заглушен
	Silenced bool

	// Immune - неуязвимость
	Immune bool

	// CanAttack - может атаковать в этот ход
	CanAttack bool

	// AttacksThisTurn - количество атак в этот ход
	AttacksThisTurn int
}

// Enchantment представляет зачарование на сущности
type Enchantment struct {
	// ID - ID зачарования
	ID string

	// EntityID - ID сущности зачарования
	EntityID int

	// AttackBonus - бонус к атаке
	AttackBonus int

	// HealthBonus - бонус к здоровью
	HealthBonus int

	// Source - источник зачарования
	Source string
}

// New создает новую сущность
func New(id int, cardID string, owner types.PlayerType) *Entity {
	return &Entity{
		ID:           id,
		CardID:       cardID,
		Owner:        owner,
		Position:     -1,
		Attributes:   Attributes{},
		Enchantments: make([]*Enchantment, 0),
	}
}

// IsAlive проверяет, жива ли сущность
func (e *Entity) IsAlive() bool {
	return e.Health > 0
}

// IsDead проверяет, мертва ли сущность
func (e *Entity) IsDead() bool {
	return e.Health <= 0
}

// CanAttackNow проверяет, может ли сущность атаковать сейчас
func (e *Entity) CanAttackNow() bool {
	if !e.Attributes.CanAttack {
		return false
	}
	if e.Attributes.Frozen {
		return false
	}
	if e.Attributes.Windfury {
		return e.Attributes.AttacksThisTurn < 2
	}
	return e.Attributes.AttacksThisTurn < 1
}

// TakeDamage наносит урон сущности
func (e *Entity) TakeDamage(damage int) {
	if e.Attributes.Immune {
		return
	}
	if e.Attributes.DivineShield && damage > 0 {
		e.Attributes.DivineShield = false
		return
	}
	e.Health -= damage
	if e.Health < 0 {
		e.Health = 0
	}
}

// Heal исцеляет сущность
func (e *Entity) Heal(amount int) {
	e.Health += amount
	if e.Health > e.MaxHealth {
		e.Health = e.MaxHealth
	}
}

// SetStats устанавливает характеристики сущности
func (e *Entity) SetStats(attack, health int) {
	e.Attack = attack
	e.Health = health
	e.MaxHealth = health
}

// ModifyAttack изменяет атаку сущности
func (e *Entity) ModifyAttack(delta int) {
	e.Attack += delta
	if e.Attack < 0 {
		e.Attack = 0
	}
}

// ModifyHealth изменяет здоровье сущности
func (e *Entity) ModifyHealth(delta int) {
	e.Health += delta
	e.MaxHealth += delta
	if e.Health < 0 {
		e.Health = 0
	}
}

// AddEnchantment добавляет зачарование к сущности
func (e *Entity) AddEnchantment(enchantment *Enchantment) {
	e.Enchantments = append(e.Enchantments, enchantment)
	e.Attack += enchantment.AttackBonus
	e.Health += enchantment.HealthBonus
	e.MaxHealth += enchantment.HealthBonus
}

// RemoveEnchantment удаляет зачарование с сущности
func (e *Entity) RemoveEnchantment(enchantmentID int) {
	for i, ench := range e.Enchantments {
		if ench.EntityID == enchantmentID {
			e.Attack -= ench.AttackBonus
			e.Health -= ench.HealthBonus
			e.MaxHealth -= ench.HealthBonus
			e.Enchantments = append(e.Enchantments[:i], e.Enchantments[i+1:]...)
			break
		}
	}
}

// Silence заглушает сущность (удаляет все эффекты)
func (e *Entity) Silence() {
	e.Attributes.Silenced = true
	e.Attributes.Taunt = false
	e.Attributes.DivineShield = false
	e.Attributes.Stealth = false
	e.Attributes.Windfury = false
	e.Attributes.Lifesteal = false
	e.Attributes.Poisonous = false

	// Удаляем все зачарования
	for _, ench := range e.Enchantments {
		e.Attack -= ench.AttackBonus
		e.Health -= ench.HealthBonus
		e.MaxHealth -= ench.HealthBonus
	}
	e.Enchantments = make([]*Enchantment, 0)
}

// ResetAttacksThisTurn сбрасывает счетчик атак в этот ход
func (e *Entity) ResetAttacksThisTurn() {
	e.Attributes.AttacksThisTurn = 0
	e.Attributes.CanAttack = true
}

// PerformAttack выполняет атаку
func (e *Entity) PerformAttack() {
	e.Attributes.AttacksThisTurn++
}
