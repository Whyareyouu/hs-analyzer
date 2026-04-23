package carddb

import "strings"

// CardType определяет тип карты.
type CardType string

const (
	CardTypeMinion    CardType = "MINION"
	CardTypeSpell     CardType = "SPELL"
	CardTypeWeapon    CardType = "WEAPON"
	CardTypeHero      CardType = "HERO"
	CardTypeHeroPower CardType = "HERO_POWER"
	CardTypeLocation  CardType = "LOCATION"
	CardTypeUnknown   CardType = "UNKNOWN"
)

// Rarity определяет редкость карты.
type Rarity string

const (
	RarityFree      Rarity = "FREE"
	RarityCommon    Rarity = "COMMON"
	RarityRare      Rarity = "RARE"
	RarityEpic      Rarity = "EPIC"
	RarityLegendary Rarity = "LEGENDARY"
	RarityUnknown   Rarity = "UNKNOWN"
)

// CardClass определяет класс карты.
type CardClass string

const (
	CardClassNeutral   CardClass = "NEUTRAL"
	CardClassDruid     CardClass = "DRUID"
	CardClassHunter    CardClass = "HUNTER"
	CardClassMage      CardClass = "MAGE"
	CardClassPaladin   CardClass = "PALADIN"
	CardClassPriest    CardClass = "PRIEST"
	CardClassRogue     CardClass = "ROGUE"
	CardClassShaman    CardClass = "SHAMAN"
	CardClassWarlock   CardClass = "WARLOCK"
	CardClassWarrior   CardClass = "WARRIOR"
	CardClassDemonHunter CardClass = "DEMONHUNTER"
	CardClassDeathKnight CardClass = "DEATHKNIGHT"
	CardClassUnknown   CardClass = "UNKNOWN"
)

// Race определяет расу сущности.
type Race string

const (
	RaceNone     Race = "NONE"
	RaceBeast    Race = "BEAST"
	RaceDemon    Race = "DEMON"
	RaceDragon   Race = "DRAGON"
	RaceElemental Race = "ELEMENTAL"
	RaceMech     Race = "MECH"
	RaceMurloc   Race = "MURLOC"
	RaceNaga     Race = "NAGA"
	RacePirate   Race = "PIRATE"
	RaceQuilboar Race = "QUILBOAR"
	RaceTotem    Race = "TOTEM"
	RaceUndead   Race = "UNDEAD"
	RaceUnknown  Race = "UNKNOWN"
)

// CardInfo содержит информацию о карте из базы данных.
type CardInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	NameRu      string    `json:"nameRu,omitempty"`
	Text        string    `json:"text,omitempty"`
	TextRu      string    `json:"textRu,omitempty"`
	Cost        int       `json:"cost,omitempty"`
	Attack      int       `json:"attack,omitempty"`
	Health      int       `json:"health,omitempty"`
	Type        CardType  `json:"type,omitempty"`
	Rarity      Rarity    `json:"rarity,omitempty"`
	Class       CardClass `json:"cardClass,omitempty"`
	Race        Race      `json:"race,omitempty"`
	Set         string    `json:"set,omitempty"`
	Collectible bool      `json:"collectible,omitempty"`
}

// GetLocalizedName возвращает локализованное имя карты.
func (c *CardInfo) GetLocalizedName(locale string) string {
	if c == nil {
		return ""
	}

	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "ru", "ru-ru", "ru_ru":
		if c.NameRu != "" {
			return c.NameRu
		}
	}
	return c.Name
}
