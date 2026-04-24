package types

// PlayerType определяет тип игрока
type PlayerType string

const (
	PlayerTypePlayer   PlayerType = "PLAYER"
	PlayerTypeOpponent PlayerType = "OPPONENT"
	PlayerTypeUnknown  PlayerType = "UNKNOWN"
)

// String возвращает строковое представление типа игрока
func (p PlayerType) String() string {
	return string(p)
}

// IsValid проверяет валидность типа игрока
func (p PlayerType) IsValid() bool {
	switch p {
	case PlayerTypePlayer, PlayerTypeOpponent, PlayerTypeUnknown:
		return true
	default:
		return false
	}
}

// ZoneType определяет тип игровой зоны
type ZoneType string

const (
	ZoneHand      ZoneType = "HAND"
	ZoneDeck      ZoneType = "DECK"
	ZonePlay      ZoneType = "PLAY"
	ZoneGraveyard ZoneType = "GRAVEYARD"
	ZoneSecret    ZoneType = "SECRET"
	ZoneSetAside  ZoneType = "SETASIDE"
	ZoneUnknown   ZoneType = "UNKNOWN"
)

// String возвращает строковое представление типа зоны
func (z ZoneType) String() string {
	return string(z)
}

// IsValid проверяет валидность типа зоны
func (z ZoneType) IsValid() bool {
	switch z {
	case ZoneHand, ZoneDeck, ZonePlay, ZoneGraveyard, ZoneSecret, ZoneSetAside, ZoneUnknown:
		return true
	default:
		return false
	}
}

// CardType определяет тип карты
type CardType string

const (
	CardTypeMinion    CardType = "MINION"
	CardTypeSpell     CardType = "SPELL"
	CardTypeWeapon    CardType = "WEAPON"
	CardTypeHero      CardType = "HERO"
	CardTypeHeroPower CardType = "HERO_POWER"
	CardTypeUnknown   CardType = "UNKNOWN"
)

// String возвращает строковое представление типа карты
func (c CardType) String() string {
	return string(c)
}

// IsValid проверяет валидность типа карты
func (c CardType) IsValid() bool {
	switch c {
	case CardTypeMinion, CardTypeSpell, CardTypeWeapon, CardTypeHero, CardTypeHeroPower, CardTypeUnknown:
		return true
	default:
		return false
	}
}

// EventType определяет тип игрового события
type EventType string

const (
	EventTypeGameStart       EventType = "game_start"
	EventTypeTurnStart       EventType = "turn_start"
	EventTypeTurnEnd         EventType = "turn_end"
	EventTypeDeckList        EventType = "deck_list"
	EventTypeCardPlayed      EventType = "card_played"
	EventTypeCardDrawn       EventType = "card_drawn"
	EventTypeAttack          EventType = "attack"
	EventTypeDamage          EventType = "damage"
	EventTypeHeal            EventType = "heal"
	EventTypeDeath           EventType = "death"
	EventTypeEffectTriggered EventType = "effect_triggered"
	EventTypeDiscover        EventType = "discover"
	EventTypeGameEnd         EventType = "game_end"
)

// String возвращает строковое представление типа события
func (e EventType) String() string {
	return string(e)
}

// IsValid проверяет валидность типа события
func (e EventType) IsValid() bool {
	switch e {
	case EventTypeGameStart, EventTypeTurnStart, EventTypeTurnEnd,
		EventTypeDeckList, EventTypeCardPlayed, EventTypeCardDrawn, EventTypeAttack,
		EventTypeDamage, EventTypeHeal, EventTypeDeath,
		EventTypeEffectTriggered, EventTypeDiscover, EventTypeGameEnd:
		return true
	default:
		return false
	}
}
