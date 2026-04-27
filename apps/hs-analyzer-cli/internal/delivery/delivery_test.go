package delivery

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"hsanalyzer/pkg/contracts/matchupload"
)

type stubUploader struct {
	result UploadResult
	err    error
	keys   []string
}

func (u *stubUploader) Upload(_ context.Context, _ matchupload.Payload, key string) (UploadResult, error) {
	u.keys = append(u.keys, key)
	return u.result, u.err
}

func TestDeliveryBufferWritesPayloadBeforeUpload(t *testing.T) {
	t.Parallel()

	spool := newTestSpool(t)
	payload := deliveryPayload()

	item, err := spool.Enqueue(payload)
	if err != nil {
		t.Fatalf("enqueue payload: %v", err)
	}

	pending, err := spool.Pending()
	if err != nil {
		t.Fatalf("load pending: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected one pending item, got %d", len(pending))
	}
	if pending[0].SpoolID != item.SpoolID || pending[0].Payload.MatchID != payload.MatchID {
		t.Fatalf("unexpected pending item: %#v", pending[0])
	}
}

func TestDeliveryBufferClearsDeliveredItemAfterSuccess(t *testing.T) {
	t.Parallel()

	spool := newTestSpool(t)
	payload := deliveryPayload()
	if _, err := spool.Enqueue(payload); err != nil {
		t.Fatalf("enqueue payload: %v", err)
	}

	uploader := &stubUploader{result: UploadResult{StatusCode: http.StatusAccepted}}
	results, err := spool.Deliver(context.Background(), uploader)
	if err != nil {
		t.Fatalf("deliver payload: %v", err)
	}
	if len(results) != 1 || results[0].Status != StatusDelivered {
		t.Fatalf("expected delivered result, got %#v", results)
	}
	if len(uploader.keys) != 1 || uploader.keys[0] != payload.IdempotencyKey() {
		t.Fatalf("unexpected idempotency keys: %#v", uploader.keys)
	}

	pending, err := spool.Pending()
	if err != nil {
		t.Fatalf("load pending: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected no pending items after delivery, got %d", len(pending))
	}
}

func TestDeliveryBufferRetriesWhenServerUnavailable(t *testing.T) {
	t.Parallel()

	spool := newTestSpool(t)
	if _, err := spool.Enqueue(deliveryPayload()); err != nil {
		t.Fatalf("enqueue payload: %v", err)
	}

	uploader := &stubUploader{
		result: UploadResult{StatusCode: http.StatusServiceUnavailable},
		err:    errors.New("server unavailable"),
	}
	results, err := spool.Deliver(context.Background(), uploader)
	if err != nil {
		t.Fatalf("deliver payload: %v", err)
	}
	if len(results) != 1 || results[0].Status != StatusRetryWait {
		t.Fatalf("expected retry_wait result, got %#v", results)
	}

	pending, err := spool.Pending()
	if err != nil {
		t.Fatalf("load pending: %v", err)
	}
	if len(pending) != 1 || pending[0].AttemptCount != 1 {
		t.Fatalf("expected retryable pending item, got %#v", pending)
	}
}

func TestDeliveryBufferMarksValidationErrorRejected(t *testing.T) {
	t.Parallel()

	spool := newTestSpool(t)
	if _, err := spool.Enqueue(deliveryPayload()); err != nil {
		t.Fatalf("enqueue payload: %v", err)
	}

	uploader := &stubUploader{result: UploadResult{StatusCode: http.StatusUnprocessableEntity}}
	results, err := spool.Deliver(context.Background(), uploader)
	if err != nil {
		t.Fatalf("deliver payload: %v", err)
	}
	if len(results) != 1 || results[0].Status != StatusRejected {
		t.Fatalf("expected rejected result, got %#v", results)
	}
}

func TestDeliveryBufferSurvivesClientRestart(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := fixedDeliveryTime()
	first, err := NewFileSpool(dir, now)
	if err != nil {
		t.Fatalf("create first spool: %v", err)
	}
	if _, err := first.Enqueue(deliveryPayload()); err != nil {
		t.Fatalf("enqueue payload: %v", err)
	}

	second, err := NewFileSpool(dir, now)
	if err != nil {
		t.Fatalf("create second spool: %v", err)
	}
	pending, err := second.Pending()
	if err != nil {
		t.Fatalf("load pending after restart: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected pending item after restart, got %d", len(pending))
	}
}

func newTestSpool(t *testing.T) *FileSpool {
	t.Helper()
	spool, err := NewFileSpool(t.TempDir(), fixedDeliveryTime())
	if err != nil {
		t.Fatalf("create spool: %v", err)
	}
	return spool
}

func fixedDeliveryTime() func() time.Time {
	current := time.Unix(100, 0).UTC()
	return func() time.Time {
		current = current.Add(time.Second)
		return current
	}
}

func deliveryPayload() matchupload.Payload {
	return matchupload.Payload{
		SchemaVersion:    matchupload.SchemaVersionV1,
		MatchID:          "match-delivery",
		CompletionStatus: matchupload.CompletionStatusComplete,
		CompletionReason: matchupload.CompletionReasonGameEnd,
		Source:           matchupload.Source{ClientName: "hs-analyzer-cli"},
		Events: []matchupload.Event{
			{Index: 0, Type: "game_start"},
		},
	}
}
