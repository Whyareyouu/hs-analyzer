package formatter

import "hsanalyzer/internal/state"

func extractMatchStates(gameState *state.State) []*state.State {
	if gameState == nil {
		return nil
	}

	if raw := gameState.GetMetadata("matches"); raw != nil {
		if matches, ok := raw.([]*state.State); ok && len(matches) > 0 {
			return matches
		}
	}

	return []*state.State{gameState}
}

func wrapMatchDocuments(mode string, matches []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"schema_version": "0.0.1",
		"mode":           mode,
		"match_count":    len(matches),
		"matches":        matches,
	}
}
