package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"hsanalyzer/pkg/contracts/matchupload"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusSending   Status = "sending"
	StatusDelivered Status = "delivered"
	StatusRetryWait Status = "retry_wait"
	StatusRejected  Status = "rejected"
)

type Item struct {
	SpoolID        string              `json:"spool_id"`
	MatchID        string              `json:"match_id"`
	IdempotencyKey string              `json:"idempotency_key"`
	PayloadHash    string              `json:"payload_hash"`
	Status         Status              `json:"status"`
	AttemptCount   int                 `json:"attempt_count"`
	LastAttemptAt  string              `json:"last_attempt_at,omitempty"`
	NextRetryAt    string              `json:"next_retry_at,omitempty"`
	LastError      string              `json:"last_error,omitempty"`
	CreatedAt      string              `json:"created_at"`
	UpdatedAt      string              `json:"updated_at"`
	Payload        matchupload.Payload `json:"payload"`
}

type UploadResult struct {
	StatusCode int
	Response   matchupload.Response
}

type Uploader interface {
	Upload(ctx context.Context, payload matchupload.Payload, idempotencyKey string) (UploadResult, error)
}

type HTTPUploader struct {
	Endpoint string
	Client   *http.Client
}

func (u HTTPUploader) Upload(ctx context.Context, payload matchupload.Payload, idempotencyKey string) (UploadResult, error) {
	if strings.TrimSpace(u.Endpoint) == "" {
		return UploadResult{}, errors.New("upload endpoint is empty")
	}
	if err := matchupload.ValidateIdempotencyKey(idempotencyKey); err != nil {
		return UploadResult{}, err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return UploadResult{}, fmt.Errorf("marshal upload payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.Endpoint, bytes.NewReader(body))
	if err != nil {
		return UploadResult{}, fmt.Errorf("create upload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(matchupload.HTTPHeaderIdempotencyKey, idempotencyKey)

	client := u.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return UploadResult{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	result := UploadResult{StatusCode: resp.StatusCode}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		_ = json.NewDecoder(resp.Body).Decode(&result.Response)
		return result, nil
	}

	return result, fmt.Errorf("upload failed with status %d", resp.StatusCode)
}

type FileSpool struct {
	dir string
	now func() time.Time
}

func NewFileSpool(dir string, now func() time.Time) (*FileSpool, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, errors.New("spool directory is empty")
	}
	if now == nil {
		now = time.Now
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create spool directory: %w", err)
	}
	return &FileSpool{dir: filepath.Clean(dir), now: now}, nil
}

func (s *FileSpool) Enqueue(payload matchupload.Payload) (Item, error) {
	if s == nil {
		return Item{}, errors.New("file spool is nil")
	}
	if err := payload.Validate(); err != nil {
		return Item{}, err
	}

	timestamp := s.now().UTC().Format(time.RFC3339Nano)
	item := Item{
		SpoolID:        payload.Hash(),
		MatchID:        payload.MatchID,
		IdempotencyKey: payload.IdempotencyKey(),
		PayloadHash:    payload.Hash(),
		Status:         StatusPending,
		CreatedAt:      timestamp,
		UpdatedAt:      timestamp,
		Payload:        payload,
	}
	if err := s.writeItem(item); err != nil {
		return Item{}, err
	}
	return item, nil
}

func (s *FileSpool) Pending() ([]Item, error) {
	if s == nil {
		return nil, errors.New("file spool is nil")
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read spool directory: %w", err)
	}

	items := make([]Item, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		item, err := s.readItem(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		if item.Status == StatusDelivered || item.Status == StatusRejected {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt < items[j].CreatedAt
	})
	return items, nil
}

func (s *FileSpool) Deliver(ctx context.Context, uploader Uploader) ([]Item, error) {
	if uploader == nil {
		return nil, errors.New("uploader is nil")
	}

	pending, err := s.Pending()
	if err != nil {
		return nil, err
	}

	results := make([]Item, 0, len(pending))
	for _, item := range pending {
		item.Status = StatusSending
		item.AttemptCount++
		item.LastAttemptAt = s.now().UTC().Format(time.RFC3339Nano)
		item.UpdatedAt = item.LastAttemptAt
		if err := s.writeItem(item); err != nil {
			return results, err
		}

		result, uploadErr := uploader.Upload(ctx, item.Payload, item.IdempotencyKey)
		item.UpdatedAt = s.now().UTC().Format(time.RFC3339Nano)
		switch {
		case uploadErr == nil && (result.StatusCode == http.StatusOK || result.StatusCode == http.StatusAccepted):
			item.Status = StatusDelivered
			item.LastError = ""
		case matchupload.IsRetryableStatus(result.StatusCode) || result.StatusCode == 0:
			item.Status = StatusRetryWait
			item.LastError = uploadErrorString(uploadErr, result.StatusCode)
			item.NextRetryAt = s.now().UTC().Add(time.Second).Format(time.RFC3339Nano)
		default:
			item.Status = StatusRejected
			item.LastError = uploadErrorString(uploadErr, result.StatusCode)
		}

		if err := s.writeItem(item); err != nil {
			return results, err
		}
		results = append(results, item)
	}

	return results, nil
}

func (s *FileSpool) writeItem(item Item) error {
	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal spool item: %w", err)
	}
	path := filepath.Join(s.dir, item.SpoolID+".json")
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("write temporary spool item: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace spool item: %w", err)
	}
	return nil
}

func (s *FileSpool) readItem(path string) (Item, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Item{}, fmt.Errorf("read spool item: %w", err)
	}
	var item Item
	if err := json.Unmarshal(data, &item); err != nil {
		return Item{}, fmt.Errorf("decode spool item: %w", err)
	}
	return item, nil
}

func uploadErrorString(err error, statusCode int) string {
	if err != nil {
		return err.Error()
	}
	if statusCode != 0 {
		return fmt.Sprintf("status %d", statusCode)
	}
	return ""
}
