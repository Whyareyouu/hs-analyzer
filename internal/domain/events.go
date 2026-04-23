package domain

type EventKind string

const (
	EventTurnStart EventKind = "TurnStart"
	EventDraw      EventKind = "Draw"
	EventPlayCard  EventKind = "PlayCard"
	EventDiscover  EventKind = "Discover"
	EventUnknown   EventKind = "Unknown"
)

type GameEvent interface {
	Kind() EventKind
	PlayerID() string
	Raw() RawEvent
}

type TurnStartEvent struct {
	Player   string `json:"player"`
	Turn     int    `json:"turn"`
	RawEvent RawEvent
}

func (e TurnStartEvent) Kind() EventKind   { return EventTurnStart }
func (e TurnStartEvent) PlayerID() string  { return e.Player }
func (e TurnStartEvent) Raw() RawEvent     { return e.RawEvent }

type DrawEvent struct {
	Player   string `json:"player"`
	Card     string `json:"card"`
	RawEvent RawEvent
}

func (e DrawEvent) Kind() EventKind  { return EventDraw }
func (e DrawEvent) PlayerID() string { return e.Player }
func (e DrawEvent) Raw() RawEvent    { return e.RawEvent }

type PlayCardEvent struct {
	Player   string `json:"player"`
	Card     string `json:"card"`
	RawEvent RawEvent
}

func (e PlayCardEvent) Kind() EventKind  { return EventPlayCard }
func (e PlayCardEvent) PlayerID() string { return e.Player }
func (e PlayCardEvent) Raw() RawEvent    { return e.RawEvent }

type DiscoverEvent struct {
	Player   string   `json:"player"`
	Options  []string `json:"options"`
	Chosen   string   `json:"chosen"`
	RawEvent RawEvent
}

func (e DiscoverEvent) Kind() EventKind  { return EventDiscover }
func (e DiscoverEvent) PlayerID() string { return e.Player }
func (e DiscoverEvent) Raw() RawEvent    { return e.RawEvent }

type UnknownEvent struct {
	Player   string `json:"player"`
	Reason   string `json:"reason"`
	RawEvent RawEvent
}

func (e UnknownEvent) Kind() EventKind  { return EventUnknown }
func (e UnknownEvent) PlayerID() string { return e.Player }
func (e UnknownEvent) Raw() RawEvent    { return e.RawEvent }
