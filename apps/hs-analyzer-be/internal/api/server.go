package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"hsanalyzer/pkg/contracts/matchupload"
)

var ErrDuplicateConflict = errors.New("duplicate idempotency key with different payload")

type StoredMatch struct {
	Payload        matchupload.Payload
	IdempotencyKey string
	PayloadHash    string
}

type MatchStore interface {
	SaveMatch(payload matchupload.Payload, idempotencyKey string) (stored bool, err error)
	GetMatch(matchID string) (StoredMatch, bool)
	ListMatches() []StoredMatch
}

type Server struct {
	store          MatchStore
	dashboardStore DashboardStore
}

func NewServer(store MatchStore, dashboardStores ...DashboardStore) *Server {
	server := &Server{store: store}
	if len(dashboardStores) > 0 {
		server.dashboardStore = dashboardStores[0]
	}
	return server
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/matches", s.handleMatches)
	mux.HandleFunc("/api/v1/decks", s.handleDecks)
	mux.HandleFunc("/api/v1/decks/", s.handleDeckDetails)
	return mux
}

func (s *Server) handleMatches(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.postMatch(w, r)
		return
	}
	if r.Method == http.MethodGet {
		s.listMatches(w, r)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, matchupload.ErrorValidationFailed, "method is not allowed")
}

func (s *Server) postMatch(w http.ResponseWriter, r *http.Request) {
	if s == nil || s.store == nil {
		writeError(w, http.StatusInternalServerError, matchupload.ErrorInternal, "match store is not configured")
		return
	}

	idempotencyKey := strings.TrimSpace(r.Header.Get(matchupload.HTTPHeaderIdempotencyKey))
	if err := matchupload.ValidateIdempotencyKey(idempotencyKey); err != nil {
		writeError(w, http.StatusUnprocessableEntity, matchupload.ErrorValidationFailed, err.Error())
		return
	}

	var payload matchupload.Payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusUnprocessableEntity, matchupload.ErrorValidationFailed, "invalid json payload")
		return
	}
	if err := payload.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, matchupload.ErrorValidationFailed, err.Error())
		return
	}

	stored, err := s.store.SaveMatch(payload, idempotencyKey)
	if errors.Is(err, ErrDuplicateConflict) {
		writeError(w, http.StatusConflict, matchupload.ErrorDuplicateConflict, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, matchupload.ErrorInternal, err.Error())
		return
	}

	status := matchupload.ResultStored
	statusCode := http.StatusOK
	if stored {
		status = matchupload.ResultAccepted
		statusCode = http.StatusAccepted
	}

	writeJSON(w, statusCode, matchupload.Response{
		Status:  status,
		MatchID: payload.MatchID,
	})
}

func (s *Server) listMatches(w http.ResponseWriter, _ *http.Request) {
	if s == nil || s.store == nil {
		writeError(w, http.StatusInternalServerError, matchupload.ErrorInternal, "match store is not configured")
		return
	}
	writeJSON(w, http.StatusOK, s.store.ListMatches())
}

func (s *Server) handleDecks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, matchupload.ErrorValidationFailed, "method is not allowed")
		return
	}
	if s == nil || s.dashboardStore == nil {
		writeError(w, http.StatusServiceUnavailable, matchupload.ErrorInternal, "dashboard store is not configured")
		return
	}
	writeJSON(w, http.StatusOK, s.dashboardStore.ListDecks())
}

func (s *Server) handleDeckDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, matchupload.ErrorValidationFailed, "method is not allowed")
		return
	}
	if s == nil || s.dashboardStore == nil {
		writeError(w, http.StatusServiceUnavailable, matchupload.ErrorInternal, "dashboard store is not configured")
		return
	}

	path := strings.TrimPrefix(strings.TrimSpace(r.URL.Path), "/api/v1/decks/")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		writeError(w, http.StatusNotFound, matchupload.ErrorValidationFailed, "deck endpoint not found")
		return
	}

	deckID := strings.TrimSpace(parts[0])
	resource := strings.TrimSpace(parts[1])
	switch resource {
	case "overview":
		overview, ok := s.dashboardStore.GetDeckOverview(deckID)
		if !ok {
			writeError(w, http.StatusNotFound, matchupload.ErrorValidationFailed, "deck overview not found")
			return
		}
		writeJSON(w, http.StatusOK, overview)
	case "cards":
		cards, ok := s.dashboardStore.GetDeckCards(deckID)
		if !ok {
			writeError(w, http.StatusNotFound, matchupload.ErrorValidationFailed, "deck cards not found")
			return
		}
		writeJSON(w, http.StatusOK, cards)
	case "matchups":
		matchups, ok := s.dashboardStore.GetDeckMatchups(deckID)
		if !ok {
			writeError(w, http.StatusNotFound, matchupload.ErrorValidationFailed, "deck matchups not found")
			return
		}
		writeJSON(w, http.StatusOK, matchups)
	default:
		writeError(w, http.StatusNotFound, matchupload.ErrorValidationFailed, "deck endpoint not found")
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, statusCode int, code, message string) {
	writeJSON(w, statusCode, matchupload.ErrorResponse{
		Error: matchupload.APIError{
			Code:    code,
			Message: message,
		},
	})
}
