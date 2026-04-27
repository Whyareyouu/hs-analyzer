package matchupload

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	SchemaVersionV1 = "match_upload/v1"

	CompletionStatusComplete   = "complete"
	CompletionStatusIncomplete = "incomplete"

	CompletionReasonGameEnd       = "game_end"
	CompletionReasonNextGameStart = "next_game_start"
	CompletionReasonFlush         = "flush"
	CompletionReasonStreamEnd     = "stream_end"

	HTTPHeaderIdempotencyKey = "Idempotency-Key"

	ResultAccepted          = "accepted"
	ResultStored            = "stored"
	ResultDuplicateConflict = "duplicate_conflict"
	ResultValidationFailed  = "validation_failed"

	ErrorValidationFailed  = "validation_failed"
	ErrorDuplicateConflict = "duplicate_conflict"
	ErrorServerUnavailable = "server_unavailable"
	ErrorInternal          = "internal_error"
	ErrorUnauthorized      = "unauthorized"
	ErrorForbidden         = "forbidden"
	ErrorRateLimited       = "rate_limited"
)

// Payload is the versioned wire contract sent by hs-analyzer-cli to hs-analyzer-be.
type Payload struct {
	SchemaVersion    string                 `json:"schema_version"`
	MatchID          string                 `json:"match_id"`
	GameID           string                 `json:"game_id,omitempty"`
	StartedAt        string                 `json:"started_at,omitempty"`
	EndedAt          string                 `json:"ended_at,omitempty"`
	CompletionStatus string                 `json:"completion_status"`
	CompletionReason string                 `json:"completion_reason"`
	Source           Source                 `json:"source"`
	Metadata         map[string]string      `json:"metadata,omitempty"`
	Events           []Event                `json:"events"`
	Snapshots        []json.RawMessage      `json:"snapshots,omitempty"`
	Extensions       map[string]interface{} `json:"extensions,omitempty"`
}

type Source struct {
	ClientName      string `json:"client_name"`
	ClientVersion   string `json:"client_version,omitempty"`
	LogsFingerprint string `json:"logs_fingerprint,omitempty"`
}

type Event struct {
	Index     int             `json:"index"`
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type Response struct {
	Status        string `json:"status"`
	MatchID       string `json:"match_id,omitempty"`
	ServerMatchID string `json:"server_match_id,omitempty"`
	Message       string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

func (p Payload) Validate() error {
	var details []ErrorDetail

	if strings.TrimSpace(p.SchemaVersion) != SchemaVersionV1 {
		details = append(details, ErrorDetail{Field: "schema_version", Message: "unsupported schema version"})
	}
	if strings.TrimSpace(p.MatchID) == "" {
		details = append(details, ErrorDetail{Field: "match_id", Message: "match id is required"})
	}
	if !isAllowedCompletionStatus(p.CompletionStatus) {
		details = append(details, ErrorDetail{Field: "completion_status", Message: "invalid completion status"})
	}
	if !isAllowedCompletionReason(p.CompletionReason) {
		details = append(details, ErrorDetail{Field: "completion_reason", Message: "invalid completion reason"})
	}
	if strings.TrimSpace(p.Source.ClientName) == "" {
		details = append(details, ErrorDetail{Field: "source.client_name", Message: "client name is required"})
	}
	if p.CompletionStatus == CompletionStatusComplete && len(p.Events) == 0 {
		details = append(details, ErrorDetail{Field: "events", Message: "completed matches require events"})
	}
	for _, value := range []struct {
		field string
		raw   string
	}{
		{field: "started_at", raw: p.StartedAt},
		{field: "ended_at", raw: p.EndedAt},
	} {
		if strings.TrimSpace(value.raw) == "" {
			continue
		}
		if _, err := time.Parse(time.RFC3339Nano, value.raw); err != nil {
			details = append(details, ErrorDetail{Field: value.field, Message: "timestamp must be RFC3339Nano"})
		}
	}

	if len(details) > 0 {
		return ValidationError{Details: details}
	}
	return nil
}

func (p Payload) IdempotencyKey() string {
	if key := strings.TrimSpace(p.MatchID); key != "" {
		return key
	}
	if key := strings.TrimSpace(p.GameID); key != "" {
		return key
	}
	return p.Hash()
}

func (p Payload) Hash() string {
	normalized := p
	normalized.Extensions = nil
	data, err := json.Marshal(normalized)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func ValidateIdempotencyKey(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("idempotency key is required")
	}
	return nil
}

func IsRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

type ValidationError struct {
	Details []ErrorDetail
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("match upload validation failed: %d error(s)", len(e.Details))
}

func isAllowedCompletionStatus(value string) bool {
	switch value {
	case CompletionStatusComplete, CompletionStatusIncomplete:
		return true
	default:
		return false
	}
}

func isAllowedCompletionReason(value string) bool {
	switch value {
	case CompletionReasonGameEnd, CompletionReasonNextGameStart, CompletionReasonFlush, CompletionReasonStreamEnd:
		return true
	default:
		return false
	}
}
