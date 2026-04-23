package game

import "testing"

func TestNewEntity(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)

	if entity.ID != 42 {
		t.Errorf("Expected ID 42, got %d", entity.ID)
	}
	if entity.CardID != "CS2_189" {
		t.Errorf("Expected CardID CS2_189, got %s", entity.CardID)
	}
	if entity.Owner != PlayerTypePlayer {
		t.Errorf("Expected Owner PLAYER, got %s", entity.Owner)
	}
	if entity.Position != -1 {
		t.Errorf("Expected Position -1, got %d", entity.Position)
	}
	if len(entity.Enchantments) != 0 {
		t.Errorf("Expected 0 enchantments, got %d", len(entity.Enchantments))
	}
}

func TestEntity_IsAlive(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)

	entity.Health = 5
	if !entity.IsAlive() {
		t.Error("Entity should be alive")
	}

	entity.Health = 0
	if entity.IsAlive() {
		t.Error("Entity should not be alive")
	}

	entity.Health = -1
	if entity.IsAlive() {
		t.Error("Entity should not be alive")
	}
}

func TestEntity_IsDead(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)

	entity.Health = 5
	if entity.IsDead() {
		t.Error("Entity should not be dead")
	}

	entity.Health = 0
	if !entity.IsDead() {
		t.Error("Entity should be dead")
	}
}

func TestEntity_CanAttackNow(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)

	// Не может атаковать по умолчанию
	if entity.CanAttackNow() {
		t.Error("Entity should not be able to attack")
	}

	// Может атаковать
	entity.Attributes.CanAttack = true
	if !entity.CanAttackNow() {
		t.Error("Entity should be able to attack")
	}

	// Заморожен
	entity.Attributes.Frozen = true
	if entity.CanAttackNow() {
		t.Error("Frozen entity should not be able to attack")
	}
	entity.Attributes.Frozen = false

	// Уже атаковал
	entity.Attributes.AttacksThisTurn = 1
	if entity.CanAttackNow() {
		t.Error("Entity should not be able to attack again")
	}

	// Windfury - может атаковать дважды
	entity.Attributes.Windfury = true
	entity.Attributes.AttacksThisTurn = 1
	if !entity.CanAttackNow() {
		t.Error("Windfury entity should be able to attack again")
	}

	entity.Attributes.AttacksThisTurn = 2
	if entity.CanAttackNow() {
		t.Error("Windfury entity should not be able to attack third time")
	}
}

func TestEntity_TakeDamage(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)
	entity.Health = 10

	entity.TakeDamage(3)
	if entity.Health != 7 {
		t.Errorf("Expected Health 7, got %d", entity.Health)
	}

	// Урон больше здоровья
	entity.TakeDamage(10)
	if entity.Health != 0 {
		t.Errorf("Expected Health 0, got %d", entity.Health)
	}
}

func TestEntity_TakeDamage_DivineShield(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)
	entity.Health = 10
	entity.Attributes.DivineShield = true

	entity.TakeDamage(5)

	if entity.Health != 10 {
		t.Errorf("Expected Health 10 (divine shield absorbed damage), got %d", entity.Health)
	}
	if entity.Attributes.DivineShield {
		t.Error("Divine shield should be removed")
	}
}

func TestEntity_TakeDamage_Immune(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)
	entity.Health = 10
	entity.Attributes.Immune = true

	entity.TakeDamage(5)

	if entity.Health != 10 {
		t.Errorf("Expected Health 10 (immune to damage), got %d", entity.Health)
	}
}

func TestEntity_Heal(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)
	entity.Health = 5
	entity.MaxHealth = 10

	entity.Heal(3)
	if entity.Health != 8 {
		t.Errorf("Expected Health 8, got %d", entity.Health)
	}

	// Исцеление больше максимума
	entity.Heal(5)
	if entity.Health != 10 {
		t.Errorf("Expected Health 10 (capped at max), got %d", entity.Health)
	}
}

func TestEntity_SetStats(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)

	entity.SetStats(3, 5)

	if entity.Attack != 3 {
		t.Errorf("Expected Attack 3, got %d", entity.Attack)
	}
	if entity.Health != 5 {
		t.Errorf("Expected Health 5, got %d", entity.Health)
	}
	if entity.MaxHealth != 5 {
		t.Errorf("Expected MaxHealth 5, got %d", entity.MaxHealth)
	}
}

func TestEntity_ModifyAttack(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)
	entity.Attack = 3

	entity.ModifyAttack(2)
	if entity.Attack != 5 {
		t.Errorf("Expected Attack 5, got %d", entity.Attack)
	}

	entity.ModifyAttack(-3)
	if entity.Attack != 2 {
		t.Errorf("Expected Attack 2, got %d", entity.Attack)
	}

	// Не может быть меньше 0
	entity.ModifyAttack(-10)
	if entity.Attack != 0 {
		t.Errorf("Expected Attack 0, got %d", entity.Attack)
	}
}

func TestEntity_ModifyHealth(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)
	entity.Health = 5
	entity.MaxHealth = 5

	entity.ModifyHealth(2)
	if entity.Health != 7 {
		t.Errorf("Expected Health 7, got %d", entity.Health)
	}
	if entity.MaxHealth != 7 {
		t.Errorf("Expected MaxHealth 7, got %d", entity.MaxHealth)
	}

	entity.ModifyHealth(-10)
	if entity.Health != 0 {
		t.Errorf("Expected Health 0, got %d", entity.Health)
	}
}

func TestEntity_AddEnchantment(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)
	entity.SetStats(3, 5)

	enchantment := &Enchantment{
		ID:          "ENCH_1",
		EntityID:    100,
		AttackBonus: 2,
		HealthBonus: 3,
		Source:      "Blessing",
	}

	entity.AddEnchantment(enchantment)

	if entity.Attack != 5 {
		t.Errorf("Expected Attack 5, got %d", entity.Attack)
	}
	if entity.Health != 8 {
		t.Errorf("Expected Health 8, got %d", entity.Health)
	}
	if entity.MaxHealth != 8 {
		t.Errorf("Expected MaxHealth 8, got %d", entity.MaxHealth)
	}
	if len(entity.Enchantments) != 1 {
		t.Errorf("Expected 1 enchantment, got %d", len(entity.Enchantments))
	}
}

func TestEntity_RemoveEnchantment(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)
	entity.SetStats(3, 5)

	enchantment := &Enchantment{
		ID:          "ENCH_1",
		EntityID:    100,
		AttackBonus: 2,
		HealthBonus: 3,
		Source:      "Blessing",
	}

	entity.AddEnchantment(enchantment)
	entity.RemoveEnchantment(100)

	if entity.Attack != 3 {
		t.Errorf("Expected Attack 3, got %d", entity.Attack)
	}
	if entity.Health != 5 {
		t.Errorf("Expected Health 5, got %d", entity.Health)
	}
	if entity.MaxHealth != 5 {
		t.Errorf("Expected MaxHealth 5, got %d", entity.MaxHealth)
	}
	if len(entity.Enchantments) != 0 {
		t.Errorf("Expected 0 enchantments, got %d", len(entity.Enchantments))
	}
}

func TestEntity_Silence(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)
	entity.SetStats(3, 5)
	entity.Attributes.Taunt = true
	entity.Attributes.DivineShield = true
	entity.Attributes.Windfury = true

	enchantment := &Enchantment{
		ID:          "ENCH_1",
		EntityID:    100,
		AttackBonus: 2,
		HealthBonus: 3,
		Source:      "Blessing",
	}
	entity.AddEnchantment(enchantment)

	entity.Silence()

	if !entity.Attributes.Silenced {
		t.Error("Entity should be silenced")
	}
	if entity.Attributes.Taunt {
		t.Error("Taunt should be removed")
	}
	if entity.Attributes.DivineShield {
		t.Error("Divine Shield should be removed")
	}
	if entity.Attributes.Windfury {
		t.Error("Windfury should be removed")
	}
	if entity.Attack != 3 {
		t.Errorf("Expected Attack 3 (enchantment removed), got %d", entity.Attack)
	}
	if len(entity.Enchantments) != 0 {
		t.Errorf("Expected 0 enchantments, got %d", len(entity.Enchantments))
	}
}

func TestEntity_ResetAttacksThisTurn(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)
	entity.Attributes.AttacksThisTurn = 2
	entity.Attributes.CanAttack = false

	entity.ResetAttacksThisTurn()

	if entity.Attributes.AttacksThisTurn != 0 {
		t.Errorf("Expected AttacksThisTurn 0, got %d", entity.Attributes.AttacksThisTurn)
	}
	if !entity.Attributes.CanAttack {
		t.Error("CanAttack should be true")
	}
}

func TestEntity_PerformAttack(t *testing.T) {
	entity := NewEntity(42, "CS2_189", PlayerTypePlayer)
	entity.Attributes.AttacksThisTurn = 0

	entity.PerformAttack()

	if entity.Attributes.AttacksThisTurn != 1 {
		t.Errorf("Expected AttacksThisTurn 1, got %d", entity.Attributes.AttacksThisTurn)
	}
}
