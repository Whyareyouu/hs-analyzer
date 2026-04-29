package matchupload

import "encoding/json"

const (
	ExtensionDashboardSummary = "dashboard_summary"

	DashboardResultWin     = "win"
	DashboardResultLoss    = "loss"
	DashboardResultUnknown = "unknown"
)

type DashboardSummary struct {
	DeckID        string              `json:"deck_id,omitempty"`
	DeckName      string              `json:"deck_name,omitempty"`
	PlayerClass   string              `json:"player_class,omitempty"`
	OpponentClass string              `json:"opponent_class,omitempty"`
	Result        string              `json:"result,omitempty"`
	DeckCards     []DashboardDeckCard `json:"deck_cards,omitempty"`
}

type DashboardDeckCard struct {
	CardID      string `json:"card_id"`
	Name        string `json:"name,omitempty"`
	ManaCost    int    `json:"mana_cost,omitempty"`
	Copies      int    `json:"copies,omitempty"`
	Rarity      string `json:"rarity,omitempty"`
	IsLegendary bool   `json:"is_legendary,omitempty"`
	Image       string `json:"image,omitempty"`
}

func (p *Payload) SetDashboardSummary(summary DashboardSummary) error {
	if p == nil {
		return nil
	}

	if p.Extensions == nil {
		p.Extensions = make(map[string]interface{})
	}
	p.Extensions[ExtensionDashboardSummary] = summary
	return nil
}

func (p Payload) GetDashboardSummary() (DashboardSummary, bool, error) {
	if p.Extensions == nil {
		return DashboardSummary{}, false, nil
	}
	raw, ok := p.Extensions[ExtensionDashboardSummary]
	if !ok || raw == nil {
		return DashboardSummary{}, false, nil
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return DashboardSummary{}, false, err
	}

	var summary DashboardSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return DashboardSummary{}, false, err
	}

	return summary, true, nil
}
