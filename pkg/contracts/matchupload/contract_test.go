package matchupload

import (
	"errors"
	"net/http"
	"testing"
)

func TestMatchUploadV1AcceptsValidPayload(t *testing.T) {
	t.Parallel()

	payload := validPayload()
	if err := payload.Validate(); err != nil {
		t.Fatalf("validate payload: %v", err)
	}
	if payload.IdempotencyKey() != payload.MatchID {
		t.Fatalf("unexpected idempotency key: %s", payload.IdempotencyKey())
	}
	if payload.Hash() == "" {
		t.Fatalf("expected payload hash")
	}
}

func TestMatchUploadV1RejectsUnknownSchemaVersion(t *testing.T) {
	t.Parallel()

	payload := validPayload()
	payload.SchemaVersion = "match_upload/v0"

	err := payload.Validate()
	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestMatchUploadV1RejectsMissingRequiredFields(t *testing.T) {
	t.Parallel()

	payload := validPayload()
	payload.MatchID = ""
	payload.Source.ClientName = ""

	err := payload.Validate()
	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(validationErr.Details) < 2 {
		t.Fatalf("expected multiple validation details, got %#v", validationErr.Details)
	}
}

func TestMatchUploadV1RejectsInvalidCompletionStatus(t *testing.T) {
	t.Parallel()

	payload := validPayload()
	payload.CompletionStatus = "done"

	if err := payload.Validate(); err == nil {
		t.Fatalf("expected invalid completion status error")
	}
}

func TestMatchUploadV1RejectsMissingIdempotencyKey(t *testing.T) {
	t.Parallel()

	if err := ValidateIdempotencyKey(""); err == nil {
		t.Fatalf("expected missing idempotency key error")
	}
}

func TestMatchUploadV1ClassifiesRetryableHTTPStatuses(t *testing.T) {
	t.Parallel()

	if !IsRetryableStatus(http.StatusServiceUnavailable) {
		t.Fatalf("expected 503 to be retryable")
	}
	if IsRetryableStatus(http.StatusUnprocessableEntity) {
		t.Fatalf("expected 422 to be non-retryable")
	}
}

func validPayload() Payload {
	return Payload{
		SchemaVersion:    SchemaVersionV1,
		MatchID:          "match-1",
		StartedAt:        "2026-04-27T12:00:00Z",
		EndedAt:          "2026-04-27T12:05:00Z",
		CompletionStatus: CompletionStatusComplete,
		CompletionReason: CompletionReasonGameEnd,
		Source:           Source{ClientName: "hs-analyzer-cli"},
		Events: []Event{
			{Index: 0, Type: "game_start"},
		},
	}
}
