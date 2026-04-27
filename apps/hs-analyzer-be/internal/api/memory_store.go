package api

import (
	"sort"
	"sync"

	"hsanalyzer/pkg/contracts/matchupload"
)

type MemoryStore struct {
	mu          sync.Mutex
	byMatchID   map[string]StoredMatch
	byIdemKey   map[string]string
	payloadHash map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		byMatchID:   make(map[string]StoredMatch),
		byIdemKey:   make(map[string]string),
		payloadHash: make(map[string]string),
	}
}

func (s *MemoryStore) SaveMatch(payload matchupload.Payload, idempotencyKey string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := payload.Hash()
	if existingHash, ok := s.payloadHash[idempotencyKey]; ok {
		if existingHash != hash {
			return false, ErrDuplicateConflict
		}
		return false, nil
	}

	record := StoredMatch{
		Payload:        payload,
		IdempotencyKey: idempotencyKey,
		PayloadHash:    hash,
	}
	s.byMatchID[payload.MatchID] = record
	s.byIdemKey[idempotencyKey] = payload.MatchID
	s.payloadHash[idempotencyKey] = hash
	return true, nil
}

func (s *MemoryStore) GetMatch(matchID string) (StoredMatch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	match, ok := s.byMatchID[matchID]
	return match, ok
}

func (s *MemoryStore) ListMatches() []StoredMatch {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]StoredMatch, 0, len(s.byMatchID))
	for _, match := range s.byMatchID {
		result = append(result, match)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Payload.MatchID < result[j].Payload.MatchID
	})
	return result
}
