package entity

import (
	"testing"

	"hsanalyzer/internal/types"
)

func TestNew(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)

	if e.ID != 42 {
		t.Errorf("Expected ID 42, got %d", e.ID)
	}
	if e.CardID != "CS2_189" {
		t.Errorf("Expected CardID CS2_189, got %s", e.CardID)
	}
	if e.Owner != types.PlayerTypePlayer {
		t.Errorf("Expected Owner PLAYER, got %s", e.Owner)
	}
	if e.Position != -1 {
		t.Errorf("Expected Position -1, got %d", e.Position)
	}
	if len(e.Enchantments) != 0 {
		t.Errorf("Expected 0 enchantments, got %d", len(e.Enchantments))
	}
}

func TestEntity_IsAlive(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)

	e.Health = 5
	if !e.IsAlive() {
		t.Error("Entity should be alive")
	}

	e.Health = 0
	if e.IsAlive() {
		t.Error("Entity should not be alive")
	}

	e.Health = -1
	if e.IsAlive() {
		t.Error("Entity should not be alive")
	}
}

func TestEntity_IsDead(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)

	e.Health = 5
	if e.IsDead() {
		t.Error("Entity should not be dead")
	}

	e.Health = 0
	if !e.IsDead() {
		t.Error("Entity should be dead")
	}
}

func TestEntity_CanAttackNow(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)

	// Не может атаковать по умолчанию
	if e.CanAttackNow() {
		t.Error("Entity should not be able to attack")
	}

	// Может атаковать
	e.Attributes.CanAttack = true
	if !e.CanAttackNow() {
		t.Error("Entity should be able to attack")
	}

	// Заморожен
	e.Attributes.Frozen = true
	if e.CanAttackNow() {
		t.Error("Frozen entity should not be able to attack")
	}
	e.Attributes.Frozen = false

	// Уже атаковал
	e.Attributes.AttacksThisTurn = 1
	if e.CanAttackNow() {
		t.Error("Entity should not be able to attack again")
	}

	// Windfury - может атаковать дважды
	e.Attributes.Windfury = true
	e.Attributes.AttacksThisTurn = 1
	if !e.CanAttackNow() {
		t.Error("Windfury entity should be able to attack again")
	}

	e.Attributes.AttacksThisTurn = 2
	if e.CanAttackNow() {
		t.Error("Windfury entity should not be able to attack third time")
	}
}

func TestEntity_TakeDamage(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)
	e.Health = 10

	e.TakeDamage(3)
	if e.Health != 7 {
		t.Errorf("Expected Health 7, got %d", e.Health)
	}

	// Урон больше здоровья
	e.TakeDamage(10)
	if e.Health != 0 {
		t.Errorf("Expected Health 0, got %d", e.Health)
	}
}

func TestEntity_TakeDamage_DivineShield(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)
	e.Health = 10
	e.Attributes.DivineShield = true

	e.TakeDamage(5)

	if e.Health != 10 {
		t.Errorf("Expected Health 10 (divine shield absorbed damage), got %d", e.Health)
	}
	if e.Attributes.DivineShield {
		t.Error("Divine shield should be removed")
	}
}

func TestEntity_TakeDamage_Immune(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)
	e.Health = 10
	e.Attributes.Immune = true

	e.TakeDamage(5)

	if e.Health != 10 {
		t.Errorf("Expected Health 10 (immune to damage), got %d", e.Health)
	}
}

func TestEntity_Heal(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)
	e.Health = 5
	e.MaxHealth = 10

	e.Heal(3)
	if e.Health != 8 {
		t.Errorf("Expected Health 8, got %d", e.Health)
	}

	// Исцеление больше максимума
	e.Heal(5)
	if e.Health != 10 {
		t.Errorf("Expected Health 10 (capped at max), got %d", e.Health)
	}
}

func TestEntity_SetStats(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)

	e.SetStats(3, 5)

	if e.Attack != 3 {
		t.Errorf("Expected Attack 3, got %d", e.Attack)
	}
	if e.Health != 5 {
		t.Errorf("Expected Health 5, got %d", e.Health)
	}
	if e.MaxHealth != 5 {
		t.Errorf("Expected MaxHealth 5, got %d", e.MaxHealth)
	}
}

func TestEntity_ModifyAttack(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)
	e.Attack = 3

	e.ModifyAttack(2)
	if e.Attack != 5 {
		t.Errorf("Expected Attack 5, got %d", e.Attack)
	}

	e.ModifyAttack(-3)
	if e.Attack != 2 {
		t.Errorf("Expected Attack 2, got %d", e.Attack)
	}

	// Не может быть меньше 0
	e.ModifyAttack(-10)
	if e.Attack != 0 {
		t.Errorf("Expected Attack 0, got %d", e.Attack)
	}
}

func TestEntity_ModifyHealth(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)
	e.Health = 5
	e.MaxHealth = 5

	e.ModifyHealth(2)
	if e.Health != 7 {
		t.Errorf("Expected Health 7, got %d", e.Health)
	}
	if e.MaxHealth != 7 {
		t.Errorf("Expected MaxHealth 7, got %d", e.MaxHealth)
	}

	e.ModifyHealth(-10)
	if e.Health != 0 {
		t.Errorf("Expected Health 0, got %d", e.Health)
	}
}

func TestEntity_AddEnchantment(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)
	e.SetStats(3, 5)

	enchantment := &Enchantment{
		ID:          "ENCH_1",
		EntityID:    100,
		AttackBonus: 2,
		HealthBonus: 3,
		Source:      "Blessing",
	}

	e.AddEnchantment(enchantment)

	if e.Attack != 5 {
		t.Errorf("Expected Attack 5, got %d", e.Attack)
	}
	if e.Health != 8 {
		t.Errorf("Expected Health 8, got %d", e.Health)
	}
	if e.MaxHealth != 8 {
		t.Errorf("Expected MaxHealth 8, got %d", e.MaxHealth)
	}
	if len(e.Enchantments) != 1 {
		t.Errorf("Expected 1 enchantment, got %d", len(e.Enchantments))
	}
}

func TestEntity_RemoveEnchantment(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)
	e.SetStats(3, 5)

	enchantment := &Enchantment{
		ID:          "ENCH_1",
		EntityID:    100,
		AttackBonus: 2,
		HealthBonus: 3,
		Source:      "Blessing",
	}

	e.AddEnchantment(enchantment)
	e.RemoveEnchantment(100)

	if e.Attack != 3 {
		t.Errorf("Expected Attack 3, got %d", e.Attack)
	}
	if e.Health != 5 {
		t.Errorf("Expected Health 5, got %d", e.Health)
	}
	if e.MaxHealth != 5 {
		t.Errorf("Expected MaxHealth 5, got %d", e.MaxHealth)
	}
	if len(e.Enchantments) != 0 {
		t.Errorf("Expected 0 enchantments, got %d", len(e.Enchantments))
	}
}

func TestEntity_Silence(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)
	e.SetStats(3, 5)
	e.Attributes.Taunt = true
	e.Attributes.DivineShield = true
	e.Attributes.Windfury = true

	enchantment := &Enchantment{
		ID:          "ENCH_1",
		EntityID:    100,
		AttackBonus: 2,
		HealthBonus: 3,
		Source:      "Blessing",
	}
	e.AddEnchantment(enchantment)

	e.Silence()

	if !e.Attributes.Silenced {
		t.Error("Entity should be silenced")
	}
	if e.Attributes.Taunt {
		t.Error("Taunt should be removed")
	}
	if e.Attributes.DivineShield {
		t.Error("Divine Shield should be removed")
	}
	if e.Attributes.Windfury {
		t.Error("Windfury should be removed")
	}
	if e.Attack != 3 {
		t.Errorf("Expected Attack 3 (enchantment removed), got %d", e.Attack)
	}
	if len(e.Enchantments) != 0 {
		t.Errorf("Expected 0 enchantments, got %d", len(e.Enchantments))
	}
}

func TestEntity_ResetAttacksThisTurn(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)
	e.Attributes.AttacksThisTurn = 2
	e.Attributes.CanAttack = false

	e.ResetAttacksThisTurn()

	if e.Attributes.AttacksThisTurn != 0 {
		t.Errorf("Expected AttacksThisTurn 0, got %d", e.Attributes.AttacksThisTurn)
	}
	if !e.Attributes.CanAttack {
		t.Error("CanAttack should be true")
	}
}

func TestEntity_PerformAttack(t *testing.T) {
	e := New(42, "CS2_189", types.PlayerTypePlayer)
	e.Attributes.AttacksThisTurn = 0

	e.PerformAttack()

	if e.Attributes.AttacksThisTurn != 1 {
		t.Errorf("Expected AttacksThisTurn 1, got %d", e.Attributes.AttacksThisTurn)
	}
}
