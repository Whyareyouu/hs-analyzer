package domain

import "time"

type RawEvent struct {
	Timestamp time.Time
	Type      string                 // CREATE_GAME, FULL_ENTITY, TAG_CHANGE, etc.
	Data      map[string]interface{} // entity_id, card_id, tag, value
	RawLine   string
}