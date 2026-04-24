package formatter

import (
	"fmt"
	"time"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
)

// JSONFormatter сериализует итоговое состояние игры в JSON-хронологию по ходам.
type JSONFormatter struct {
	cardDB carddb.CardDatabase
}

// NewJSONFormatter создает JSON formatter.
func NewJSONFormatter(cardDB carddb.CardDatabase) *JSONFormatter {
	return &JSONFormatter{cardDB: cardDB}
}

func (f *JSONFormatter) Name() string {
	return "json"
}

func (f *JSONFormatter) Format(gameState *state.State) (string, error) {
	if gameState == nil {
		return "", fmt.Errorf("game state is nil")
	}

	timeline := make([]jsonTurn, 0, len(gameState.Turns))
	for _, turn := range gameState.Turns {
		if turn == nil {
			continue
		}

		item := jsonTurn{
			Number:     turn.Number,
			Player:     turn.Player.String(),
			StartedAt:  turn.StartTime.Format(time.RFC3339Nano),
			EventCount: len(turn.Events),
			Events:     make([]jsonEvent, 0, len(turn.Events)),
		}
		if turn.EndTime != nil {
			item.EndedAt = turn.EndTime.Format(time.RFC3339Nano)
		}

		for index, rawEvent := range turn.Events {
			event, ok := rawEvent.(game.Event)
			if !ok {
				continue
			}

			item.Events = append(item.Events, f.buildJSONEvent(index, event, gameState))
		}

		timeline = append(timeline, item)
	}

	encoded, err := marshalIndentedJSON(timeline)
	if err != nil {
		return "", fmt.Errorf("marshal turns json: %w", err)
	}

	return encoded, nil
}

func (f *JSONFormatter) FormatEvent(event game.Event) (string, error) {
	if event == nil {
		return "", fmt.Errorf("event is nil")
	}

	encoded, err := marshalIndentedJSON(f.buildJSONEvent(0, event, nil))
	if err != nil {
		return "", fmt.Errorf("marshal event json: %w", err)
	}

	return encoded, nil
}

type jsonTurn struct {
	Number     int         `json:"number"`
	Player     string      `json:"player"`
	StartedAt  string      `json:"started_at"`
	EndedAt    string      `json:"ended_at,omitempty"`
	EventCount int         `json:"event_count"`
	Events     []jsonEvent `json:"events"`
}

type jsonEvent struct {
	Index      int                    `json:"index"`
	Type       string                 `json:"type"`
	Timestamp  string                 `json:"timestamp"`
	Summary    string                 `json:"summary,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

func (f *JSONFormatter) buildJSONEvent(index int, event game.Event, gameState *state.State) jsonEvent {
	item := jsonEvent{
		Index:      index,
		Type:       event.Type().String(),
		Timestamp:  event.Timestamp().Format(time.RFC3339Nano),
		Attributes: make(map[string]interface{}),
	}

	switch e := event.(type) {
	case *game.GameStartEvent:
		item.Summary = "game_start"
		item.Attributes["game_id"] = e.GameID
		item.Attributes["game_mode"] = e.GameMode
	case *game.GameEndEvent:
		item.Summary = "game_end"
		item.Attributes["winner"] = e.Winner.String()
	case *game.TurnStartEvent:
		item.Summary = "turn_start"
		item.Attributes["player"] = e.Player.String()
	case *game.TurnEndEvent:
		item.Summary = "turn_end"
	case *game.DeckListEvent:
		item.Summary = "deck_list"
		item.Attributes["player"] = e.Player.String()
		item.Attributes["name"] = e.Name
		item.Attributes["deck_id"] = e.DeckID
		item.Attributes["code"] = e.Code
		item.Attributes["card_count"] = len(e.Cards)
		item.Attributes["cards"] = f.buildJSONCards(e.Cards)
	case *game.CardPlayedEvent:
		item.Summary = "card_played"
		item.Attributes["player"] = e.Player.String()
		item.Attributes["position"] = e.Position
		item.Attributes["card"] = f.buildJSONCard(e.Card)
	case *game.CardDrawnEvent:
		item.Summary = "card_drawn"
		item.Attributes["player"] = e.Player.String()
		item.Attributes["card"] = f.buildJSONCard(e.Card)
	case *game.AttackEvent:
		item.Summary = "attack"
		item.Attributes["attacker_entity_id"] = e.AttackerEntityID
		item.Attributes["target_entity_id"] = e.TargetEntityID
		item.Attributes["attacker"] = f.buildJSONEventEntityRef(gameState, e.AttackerEntityID, e.AttackerCardID, e.AttackerName)
		item.Attributes["target"] = f.buildJSONEventEntityRef(gameState, e.TargetEntityID, e.TargetCardID, e.TargetName)
	case *game.DamageEvent:
		item.Summary = "damage"
		item.Attributes["target_entity_id"] = e.TargetEntityID
		item.Attributes["amount"] = e.Amount
		item.Attributes["target"] = f.buildJSONEventEntityRef(gameState, e.TargetEntityID, e.TargetCardID, e.TargetName)
	case *game.HealEvent:
		item.Summary = "heal"
		item.Attributes["target_entity_id"] = e.TargetEntityID
		item.Attributes["amount"] = e.Amount
		item.Attributes["target"] = f.buildJSONEventEntityRef(gameState, e.TargetEntityID, e.TargetCardID, e.TargetName)
	case *game.DeathEvent:
		item.Summary = "death"
		item.Attributes["entity_id"] = e.EntityID
		item.Attributes["entity"] = f.buildJSONEventEntityRef(gameState, e.EntityID, e.CardID, e.Name)
	case *game.EffectTriggeredEvent:
		sourceName, sourceText := f.resolveEffectSource(e)
		details := describeEffect(e.EffectName, sourceName, sourceText, "ru")
		item.Summary = details.Label
		item.Attributes["source_entity_id"] = e.SourceEntityID
		item.Attributes["effect_name"] = e.EffectName
		item.Attributes["effect_label"] = details.Label
		if details.Description != "" {
			item.Attributes["effect_description"] = details.Description
		}
		if e.SourceCardID != "" {
			item.Attributes["source_card_id"] = e.SourceCardID
			item.Attributes["source_card"] = f.buildJSONCardByID(e.SourceCardID)
		}
		if sourceName != "" {
			item.Attributes["source_name"] = sourceName
		}
		if sourceText != "" {
			item.Attributes["source_card_text"] = sourceText
		}
	case *game.DiscoverEvent:
		item.Summary = "discover"
		item.Attributes["player"] = e.Player.String()
		item.Attributes["options"] = append([]string(nil), e.Options...)
		item.Attributes["chosen_card"] = e.ChosenCard
		item.Attributes["chosen_card_id"] = e.ChosenCardID
		item.Attributes["option_cards"] = f.buildJSONCardRefs(e.Options)
		if e.ChosenCard != "" {
			item.Attributes["chosen_card_info"] = f.buildJSONCardByID(e.ChosenCard)
		}
		if e.ChosenCardID != 0 {
			item.Attributes["chosen_entity"] = f.buildJSONEntityRef(gameState, e.ChosenCardID)
		}
	default:
		item.Summary = event.Type().String()
	}

	if len(item.Attributes) == 0 {
		item.Attributes = nil
	}

	return item
}

type jsonCard struct {
	ID         string        `json:"id"`
	Name       string        `json:"name,omitempty"`
	EntityID   int           `json:"entity_id"`
	Owner      string        `json:"owner,omitempty"`
	Controller string        `json:"controller,omitempty"`
	Zone       string        `json:"zone,omitempty"`
	Position   int           `json:"position"`
	CreatedBy  string        `json:"created_by,omitempty"`
	Info       *jsonCardInfo `json:"info,omitempty"`
}

type jsonCardRef struct {
	ID   string        `json:"id"`
	Name string        `json:"name,omitempty"`
	Info *jsonCardInfo `json:"info,omitempty"`
}

type jsonCardInfo struct {
	ID          string `json:"id"`
	DBFID       int    `json:"dbf_id"`
	Name        string `json:"name"`
	Text        string `json:"text"`
	Cost        int    `json:"cost"`
	Attack      int    `json:"attack"`
	Health      int    `json:"health"`
	Type        string `json:"type"`
	Rarity      string `json:"rarity"`
	CardClass   string `json:"card_class"`
	Race        string `json:"race"`
	Set         string `json:"set"`
	Collectible bool   `json:"collectible"`
}

type jsonEntityRef struct {
	EntityID  int           `json:"entity_id"`
	CardID    string        `json:"card_id,omitempty"`
	Name      string        `json:"name,omitempty"`
	Owner     string        `json:"owner,omitempty"`
	Zone      string        `json:"zone,omitempty"`
	Position  int           `json:"position,omitempty"`
	Attack    int           `json:"attack,omitempty"`
	Health    int           `json:"health,omitempty"`
	MaxHealth int           `json:"max_health,omitempty"`
	Card      *jsonCardRef  `json:"card,omitempty"`
	Info      *jsonCardInfo `json:"info,omitempty"`
}

func (f *JSONFormatter) buildJSONCard(c *card.Card) *jsonCard {
	if c == nil {
		return nil
	}

	info := f.resolveCardInfo(c)

	return &jsonCard{
		ID:         c.ID,
		Name:       f.resolveCardName(c, info),
		EntityID:   c.EntityID,
		Owner:      c.Owner.String(),
		Controller: c.Controller.String(),
		Zone:       string(c.Zone),
		Position:   c.Position,
		CreatedBy:  c.CreatedBy,
		Info:       buildJSONCardInfo(info),
	}
}

func (f *JSONFormatter) buildJSONCards(cards []*card.Card) []*jsonCard {
	if len(cards) == 0 {
		return nil
	}

	result := make([]*jsonCard, 0, len(cards))
	for _, current := range cards {
		if current == nil {
			continue
		}
		result = append(result, f.buildJSONCard(current))
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func (f *JSONFormatter) buildJSONCardRefs(ids []string) []*jsonCardRef {
	if len(ids) == 0 {
		return nil
	}

	result := make([]*jsonCardRef, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		result = append(result, f.buildJSONCardByID(id))
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func (f *JSONFormatter) buildJSONCardByID(cardID string) *jsonCardRef {
	if cardID == "" {
		return nil
	}

	info := f.resolveCardInfoByID(cardID)
	return &jsonCardRef{
		ID:   cardID,
		Name: f.resolveCardNameByID(cardID, info),
		Info: buildJSONCardInfo(info),
	}
}

func (f *JSONFormatter) resolveCardInfo(c *card.Card) *carddb.CardInfo {
	if c == nil {
		return nil
	}

	switch info := c.Info.(type) {
	case *carddb.CardInfo:
		return info
	case carddb.CardInfo:
		resolved := info
		return &resolved
	}

	if f == nil || f.cardDB == nil || c.ID == "" {
		return nil
	}

	info, ok := f.cardDB.GetCard(c.ID)
	if !ok {
		return nil
	}

	return info
}

func (f *JSONFormatter) resolveCardInfoByID(cardID string) *carddb.CardInfo {
	if f == nil || f.cardDB == nil || cardID == "" {
		return nil
	}

	info, ok := f.cardDB.GetCard(cardID)
	if !ok {
		return nil
	}

	return info
}

func (f *JSONFormatter) resolveCardName(c *card.Card, info *carddb.CardInfo) string {
	if info != nil && info.Name != "" {
		return info.Name
	}
	if c == nil {
		return ""
	}
	return c.ID
}

func (f *JSONFormatter) resolveCardNameByID(cardID string, info *carddb.CardInfo) string {
	if info != nil && info.Name != "" {
		return info.Name
	}

	return cardID
}

func buildJSONCardInfo(info *carddb.CardInfo) *jsonCardInfo {
	if info == nil {
		return nil
	}

	return &jsonCardInfo{
		ID:          info.ID,
		DBFID:       info.DBFID,
		Name:        info.Name,
		Text:        cleanCardText(info.Text),
		Cost:        info.Cost,
		Attack:      info.Attack,
		Health:      info.Health,
		Type:        string(info.Type),
		Rarity:      string(info.Rarity),
		CardClass:   string(info.Class),
		Race:        string(info.Race),
		Set:         info.Set,
		Collectible: info.Collectible,
	}
}

func (f *JSONFormatter) buildJSONEntityRef(gameState *state.State, entityID int) *jsonEntityRef {
	if entityID == 0 {
		return nil
	}

	item := &jsonEntityRef{
		EntityID: entityID,
	}

	if gameState == nil {
		return item
	}

	if c := gameState.GetCard(entityID); c != nil {
		info := f.resolveCardInfo(c)
		item.CardID = c.ID
		item.Name = f.resolveCardName(c, info)
		item.Owner = c.Owner.String()
		item.Zone = string(c.Zone)
		item.Position = c.Position
		item.Card = f.buildJSONCardByID(c.ID)
		item.Info = buildJSONCardInfo(info)
	}

	if entity := gameState.GetEntity(entityID); entity != nil {
		item.Attack = entity.Attack
		item.Health = entity.Health
		item.MaxHealth = entity.MaxHealth
		if item.CardID == "" {
			item.CardID = entity.CardID
			info := f.resolveCardInfoByID(entity.CardID)
			item.Name = f.resolveCardNameByID(entity.CardID, info)
			item.Owner = entity.Owner.String()
			item.Position = entity.Position
			item.Card = f.buildJSONCardByID(entity.CardID)
			item.Info = buildJSONCardInfo(info)
		}
	}

	return item
}

func (f *JSONFormatter) buildJSONEventEntityRef(gameState *state.State, entityID int, cardID, name string) *jsonEntityRef {
	item := f.buildJSONEntityRef(gameState, entityID)
	if item == nil {
		return nil
	}

	if item.CardID == "" {
		item.CardID = cardID
	}
	if item.Name == "" {
		item.Name = name
	}
	if item.Card == nil && cardID != "" {
		item.Card = f.buildJSONCardByID(cardID)
	}
	if item.Info == nil && cardID != "" {
		item.Info = buildJSONCardInfo(f.resolveCardInfoByID(cardID))
	}

	return item
}

func (f *JSONFormatter) resolveEffectSource(event *game.EffectTriggeredEvent) (string, string) {
	if event == nil {
		return "", ""
	}

	if f != nil && f.cardDB != nil && event.SourceCardID != "" {
		info, ok := f.cardDB.GetCard(event.SourceCardID)
		if ok && info != nil {
			return info.Name, cleanCardText(info.Text)
		}
	}

	return event.SourceName, ""
}
