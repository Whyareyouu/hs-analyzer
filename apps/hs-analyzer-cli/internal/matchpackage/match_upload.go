package matchpackage

import (
	"encoding/json"
	"strings"
	"time"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/monitor"
	"hsanalyzer/internal/types"
	"hsanalyzer/pkg/contracts/matchupload"
)

const defaultMatchUploadClientName = "hs-analyzer-cli"

// BuildMatchUploadPayload converts a finalized live match into the server wire contract.
func BuildMatchUploadPayload(match monitor.PersistedMatch, clientVersion string) matchupload.Payload {
	events := buildMatchUploadEvents(match.Events)
	startedAt, endedAt := resolveMatchTimeBounds(match.Events)

	payload := matchupload.Payload{
		SchemaVersion:    matchupload.SchemaVersionV1,
		MatchID:          strings.TrimSpace(match.Key),
		CompletionStatus: matchUploadCompletionStatus(match.Complete),
		CompletionReason: matchUploadCompletionReason(match.Reason),
		Source: matchupload.Source{
			ClientName:    defaultMatchUploadClientName,
			ClientVersion: strings.TrimSpace(clientVersion),
		},
		Metadata: map[string]string{},
		Events:   events,
	}

	if match.State != nil {
		payload.GameID = strings.TrimSpace(match.State.GameID)
		if match.State.GameMode != "" {
			payload.Metadata["game_mode"] = string(match.State.GameMode)
		}
	}
	if !startedAt.IsZero() {
		payload.StartedAt = startedAt.UTC().Format(time.RFC3339Nano)
	}
	if !endedAt.IsZero() {
		payload.EndedAt = endedAt.UTC().Format(time.RFC3339Nano)
	}
	if len(payload.Metadata) == 0 {
		payload.Metadata = nil
	}
	if payload.MatchID == "" {
		payload.MatchID = payload.Hash()
	}

	return payload
}

func buildMatchUploadEvents(events []game.Event) []matchupload.Event {
	result := make([]matchupload.Event, 0, len(events))
	for index, event := range events {
		if event == nil {
			continue
		}

		payload, _ := json.Marshal(map[string]string{
			"type": event.Type().String(),
		})
		result = append(result, matchupload.Event{
			Index:     index,
			Type:      event.Type().String(),
			Timestamp: event.Timestamp().UTC().Format(time.RFC3339Nano),
			Payload:   payload,
		})
	}
	return result
}

func resolveMatchTimeBounds(events []game.Event) (time.Time, time.Time) {
	var startedAt time.Time
	var endedAt time.Time
	for _, event := range events {
		if event == nil {
			continue
		}
		timestamp := event.Timestamp()
		if timestamp.IsZero() {
			continue
		}
		if startedAt.IsZero() || timestamp.Before(startedAt) {
			startedAt = timestamp
		}
		if endedAt.IsZero() || timestamp.After(endedAt) {
			endedAt = timestamp
		}
	}
	return startedAt, endedAt
}

func matchUploadCompletionStatus(complete bool) string {
	if complete {
		return matchupload.CompletionStatusComplete
	}
	return matchupload.CompletionStatusIncomplete
}

func matchUploadCompletionReason(reason monitor.CompletionReason) string {
	switch reason {
	case monitor.CompletionReasonGameEnd:
		return matchupload.CompletionReasonGameEnd
	case monitor.CompletionReasonNextGameStart:
		return matchupload.CompletionReasonNextGameStart
	case monitor.CompletionReasonFlush:
		return matchupload.CompletionReasonFlush
	case monitor.CompletionReasonStreamEnd:
		return matchupload.CompletionReasonStreamEnd
	default:
		return string(reason)
	}
}

func eventTypeString(eventType types.EventType) string {
	return eventType.String()
}
