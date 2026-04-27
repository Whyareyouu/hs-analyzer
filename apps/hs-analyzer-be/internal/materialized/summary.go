package materialized

import (
	"sort"

	"hsanalyzer/apps/hs-analyzer-be/internal/api"
	"hsanalyzer/pkg/contracts/matchupload"
)

type MatchSummary struct {
	MatchID          string `json:"match_id"`
	GameID           string `json:"game_id,omitempty"`
	CompletionStatus string `json:"completion_status"`
	CompletionReason string `json:"completion_reason"`
	EventCount       int    `json:"event_count"`
	GameMode         string `json:"game_mode,omitempty"`
}

func BuildMatchSummary(stored api.StoredMatch) MatchSummary {
	payload := stored.Payload
	return MatchSummary{
		MatchID:          payload.MatchID,
		GameID:           payload.GameID,
		CompletionStatus: payload.CompletionStatus,
		CompletionReason: payload.CompletionReason,
		EventCount:       len(payload.Events),
		GameMode:         payload.Metadata["game_mode"],
	}
}

func BuildHistory(matches []api.StoredMatch) []MatchSummary {
	result := make([]MatchSummary, 0, len(matches))
	for _, match := range matches {
		result = append(result, BuildMatchSummary(match))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].MatchID < result[j].MatchID
	})
	return result
}

func EventTypes(payload matchupload.Payload) []string {
	result := make([]string, 0, len(payload.Events))
	for _, event := range payload.Events {
		result = append(result, event.Type)
	}
	return result
}
