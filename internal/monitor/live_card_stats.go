package monitor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/textutil"
	"hsanalyzer/internal/types"
)

const (
	LiveCardStatsSchemaVersion = "live_card_stats/v1"

	RecoveryStatusFresh   = "fresh"
	RecoveryStatusResumed = "resumed"
	RecoveryStatusPartial = "partial"

	LiveCardStatsImageKindNone = "none"
	LiveCardStatsImageKindURL  = "url"
	LiveCardStatsImageKindFile = "file"

	LiveCardStatsMatchStatusIdle      = "idle"
	LiveCardStatsMatchStatusActive    = "active"
	LiveCardStatsMatchStatusFinalized = "finalized"
)

// LiveCardStatsImage is the client-facing image payload for one card.
type LiveCardStatsImage struct {
	Kind  string `json:"kind"`
	Value string `json:"value,omitempty"`
}

// LiveCardStatsCardPayload is the enriched client payload for one card.
type LiveCardStatsCardPayload struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Image       LiveCardStatsImage `json:"image"`
}

// LiveCardStatsCardEntry is the aggregated representation for one played card.
type LiveCardStatsCardEntry struct {
	CardID       string                   `json:"card_id"`
	PlayCount    int                      `json:"play_count"`
	FirstPlaySeq int                      `json:"first_play_seq"`
	LastPlaySeq  int                      `json:"last_play_seq"`
	Card         LiveCardStatsCardPayload `json:"card"`
}

// LiveCardStatsSideSnapshot stores stats for one side.
type LiveCardStatsSideSnapshot struct {
	TotalPlays  int                      `json:"total_plays"`
	UniqueCards int                      `json:"unique_cards"`
	Cards       []LiveCardStatsCardEntry `json:"cards"`
}

// LiveCardStatsPlayersSnapshot stores separate card arrays for self/opponent.
type LiveCardStatsPlayersSnapshot struct {
	Self     LiveCardStatsSideSnapshot `json:"self"`
	Opponent LiveCardStatsSideSnapshot `json:"opponent"`
}

// LiveCardStatsMatchSnapshot stores current match metadata.
type LiveCardStatsMatchSnapshot struct {
	Key       string `json:"key,omitempty"`
	Status    string `json:"status"`
	StartedAt string `json:"started_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// LiveCardStatsSnapshot is the client-facing snapshot contract.
type LiveCardStatsSnapshot struct {
	SchemaVersion     string                       `json:"schema_version"`
	ProjectionVersion int                          `json:"projection_version"`
	RecoveryStatus    string                       `json:"recovery_status"`
	Match             LiveCardStatsMatchSnapshot   `json:"match"`
	Players           LiveCardStatsPlayersSnapshot `json:"players"`
}

// CardAssetResolver resolves optional image payloads without baking provider
// knowledge into the projector.
type CardAssetResolver interface {
	ResolveCardAsset(cardID string, dbfID int) LiveCardStatsImage
}

// LiveCardStatsSink publishes snapshots to an external delivery surface.
type LiveCardStatsSink interface {
	Publish(snapshot LiveCardStatsSnapshot) error
}

// LiveCardStatsProjectorOptions configure the projector behavior.
type LiveCardStatsProjectorOptions struct {
	CardDB        carddb.CardDatabase
	AssetResolver CardAssetResolver
	Sink          LiveCardStatsSink
	Now           func() time.Time
}

type liveCardStatsAggregate struct {
	entry LiveCardStatsCardEntry
}

// LiveCardStatsProjector maintains a lightweight projection for the active
// match and publishes snapshots after meaningful lifecycle changes.
type LiveCardStatsProjector struct {
	cardDB        carddb.CardDatabase
	assetResolver CardAssetResolver
	sink          LiveCardStatsSink
	now           func() time.Time

	snapshot LiveCardStatsSnapshot
	self     map[string]*liveCardStatsAggregate
	opponent map[string]*liveCardStatsAggregate
}

// NoopCardAssetResolver is the default image resolver for the initial rollout.
type NoopCardAssetResolver struct{}

func (NoopCardAssetResolver) ResolveCardAsset(string, int) LiveCardStatsImage {
	return LiveCardStatsImage{Kind: LiveCardStatsImageKindNone}
}

// AtomicFileLiveCardStatsSink atomically replaces a JSON snapshot file.
type AtomicFileLiveCardStatsSink struct {
	path        string
	lastPayload []byte
}

// NewAtomicFileLiveCardStatsSink creates a file-backed snapshot sink.
func NewAtomicFileLiveCardStatsSink(path string) (*AtomicFileLiveCardStatsSink, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("live card stats path is empty")
	}

	return &AtomicFileLiveCardStatsSink{path: filepath.Clean(path)}, nil
}

// Publish writes the snapshot only when the serialized payload changed.
func (s *AtomicFileLiveCardStatsSink) Publish(snapshot LiveCardStatsSnapshot) error {
	if s == nil {
		return errors.New("live card stats sink is nil")
	}

	payload, err := marshalLiveCardStatsSnapshot(snapshot)
	if err != nil {
		return fmt.Errorf("marshal live card stats snapshot: %w", err)
	}

	if bytes.Equal(s.lastPayload, payload) {
		return nil
	}
	if len(s.lastPayload) == 0 {
		existing, err := os.ReadFile(s.path)
		switch {
		case err == nil && bytes.Equal(existing, payload):
			s.lastPayload = append([]byte(nil), payload...)
			return nil
		case err == nil:
		case os.IsNotExist(err):
		default:
			return fmt.Errorf("read existing live card stats snapshot: %w", err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create live card stats directory: %w", err)
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, payload, 0o644); err != nil {
		return fmt.Errorf("write temporary live card stats snapshot: %w", err)
	}
	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("replace live card stats snapshot: %w", err)
	}

	s.lastPayload = append(s.lastPayload[:0], payload...)
	return nil
}

// Path returns the current snapshot file path.
func (s *AtomicFileLiveCardStatsSink) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

// NewLiveCardStatsProjector creates an empty projector.
func NewLiveCardStatsProjector(options LiveCardStatsProjectorOptions) *LiveCardStatsProjector {
	now := options.Now
	if now == nil {
		now = time.Now
	}

	resolver := options.AssetResolver
	if resolver == nil {
		resolver = NoopCardAssetResolver{}
	}

	return &LiveCardStatsProjector{
		cardDB:        options.CardDB,
		assetResolver: resolver,
		sink:          options.Sink,
		now:           now,
		snapshot:      newIdleLiveCardStatsSnapshot(),
		self:          make(map[string]*liveCardStatsAggregate),
		opponent:      make(map[string]*liveCardStatsAggregate),
	}
}

// Process ingests one live event and publishes a snapshot on meaningful
// lifecycle changes.
func (p *LiveCardStatsProjector) Process(event game.Event) error {
	if p == nil || event == nil {
		return nil
	}

	switch current := event.(type) {
	case *game.GameStartEvent:
		return p.processGameStart(current)
	case *game.CardPlayedEvent:
		return p.processCardPlayed(current)
	case *game.GameEndEvent:
		return p.processGameEnd(current)
	default:
		return nil
	}
}

// Restore loads a previously persisted projection state and republishes it.
func (p *LiveCardStatsProjector) Restore(snapshot LiveCardStatsSnapshot) error {
	if p == nil {
		return errors.New("live card stats projector is nil")
	}

	normalized, err := normalizeLiveCardStatsSnapshot(snapshot)
	if err != nil {
		return err
	}

	p.snapshot = normalized
	p.self = make(map[string]*liveCardStatsAggregate)
	p.opponent = make(map[string]*liveCardStatsAggregate)
	p.restoreAggregates(p.self, normalized.Players.Self.Cards)
	p.restoreAggregates(p.opponent, normalized.Players.Opponent.Cards)

	if p.snapshot.Match.Status == LiveCardStatsMatchStatusActive {
		p.snapshot.RecoveryStatus = RecoveryStatusResumed
		p.snapshot.ProjectionVersion++
	}

	if p.sink != nil {
		if err := p.sink.Publish(p.Snapshot()); err != nil {
			return fmt.Errorf("publish restored live card stats snapshot: %w", err)
		}
	}

	return nil
}

// Snapshot returns a deep copy of the current snapshot.
func (p *LiveCardStatsProjector) Snapshot() LiveCardStatsSnapshot {
	if p == nil {
		return newIdleLiveCardStatsSnapshot()
	}

	return cloneLiveCardStatsSnapshot(p.snapshot)
}

// PersistenceSnapshot returns a deep copy suitable for monitor state storage.
func (p *LiveCardStatsProjector) PersistenceSnapshot() *LiveCardStatsSnapshot {
	if p == nil {
		return nil
	}
	if p.snapshot.Match.Status == LiveCardStatsMatchStatusIdle && p.snapshot.ProjectionVersion == 0 {
		return nil
	}

	snapshot := p.Snapshot()
	return &snapshot
}

func (p *LiveCardStatsProjector) processGameStart(event *game.GameStartEvent) error {
	if event == nil {
		return nil
	}

	now := p.now().UTC().Format(time.RFC3339Nano)
	p.self = make(map[string]*liveCardStatsAggregate)
	p.opponent = make(map[string]*liveCardStatsAggregate)
	p.snapshot = LiveCardStatsSnapshot{
		SchemaVersion:     LiveCardStatsSchemaVersion,
		ProjectionVersion: p.snapshot.ProjectionVersion + 1,
		RecoveryStatus:    RecoveryStatusFresh,
		Match: LiveCardStatsMatchSnapshot{
			Key:       buildLiveCardStatsMatchKey(event.GameID),
			Status:    LiveCardStatsMatchStatusActive,
			StartedAt: now,
			UpdatedAt: now,
		},
		Players: newEmptyLiveCardStatsPlayersSnapshot(),
	}

	return p.publishCurrentSnapshot()
}

func (p *LiveCardStatsProjector) processCardPlayed(event *game.CardPlayedEvent) error {
	if event == nil {
		return nil
	}
	if p.snapshot.Match.Status != LiveCardStatsMatchStatusActive {
		return nil
	}

	bucket, side, ok := p.resolveBucket(event.Player)
	if !ok {
		return nil
	}
	if !p.isTrackableCard(event.Card) {
		return nil
	}

	cardID := strings.TrimSpace(event.Card.ID)
	seq := p.snapshot.Players.Self.TotalPlays + p.snapshot.Players.Opponent.TotalPlays + 1
	entry := p.upsertCard(bucket, cardID, seq, event.Card)
	entry.PlayCount++
	entry.LastPlaySeq = seq
	if entry.FirstPlaySeq == 0 {
		entry.FirstPlaySeq = seq
	}
	entry.Card = p.resolveCardPayload(event.Card)

	p.snapshot.ProjectionVersion++
	p.snapshot.Match.UpdatedAt = p.now().UTC().Format(time.RFC3339Nano)
	if side == "self" {
		p.snapshot.Players.Self.TotalPlays++
	} else {
		p.snapshot.Players.Opponent.TotalPlays++
	}
	p.snapshot.Players = p.buildPlayersSnapshot()

	return p.publishCurrentSnapshot()
}

func (p *LiveCardStatsProjector) processGameEnd(*game.GameEndEvent) error {
	if p.snapshot.Match.Status != LiveCardStatsMatchStatusActive {
		return nil
	}

	p.snapshot.ProjectionVersion++
	p.snapshot.Match.Status = LiveCardStatsMatchStatusFinalized
	p.snapshot.Match.UpdatedAt = p.now().UTC().Format(time.RFC3339Nano)
	p.snapshot.Players = p.buildPlayersSnapshot()

	return p.publishCurrentSnapshot()
}

func (p *LiveCardStatsProjector) upsertCard(
	bucket map[string]*liveCardStatsAggregate,
	cardID string,
	seq int,
	playedCard *card.Card,
) *LiveCardStatsCardEntry {
	aggregate, ok := bucket[cardID]
	if !ok {
		aggregate = &liveCardStatsAggregate{
			entry: LiveCardStatsCardEntry{
				CardID:       cardID,
				FirstPlaySeq: seq,
				LastPlaySeq:  seq,
				Card:         p.resolveCardPayload(playedCard),
			},
		}
		bucket[cardID] = aggregate
	}
	return &aggregate.entry
}

func (p *LiveCardStatsProjector) resolveBucket(player types.PlayerType) (map[string]*liveCardStatsAggregate, string, bool) {
	switch player {
	case types.PlayerTypePlayer:
		return p.self, "self", true
	case types.PlayerTypeOpponent:
		return p.opponent, "opponent", true
	default:
		return nil, "", false
	}
}

func (p *LiveCardStatsProjector) isTrackableCard(current *card.Card) bool {
	if current == nil {
		return false
	}

	cardID := strings.TrimSpace(current.ID)
	if cardID == "" {
		return false
	}

	info := p.resolveCardInfo(current)
	return info == nil || info.Type != carddb.CardTypeHeroPower
}

func (p *LiveCardStatsProjector) resolveCardPayload(current *card.Card) LiveCardStatsCardPayload {
	cardID := ""
	if current != nil {
		cardID = strings.TrimSpace(current.ID)
	}

	info := p.resolveCardInfo(current)
	dbfID := 0
	name := cardID
	description := ""
	if info != nil {
		dbfID = info.DBFID
		if info.Name != "" {
			name = info.Name
		}
		description = textutil.CleanCardText(info.Text)
	}

	image := p.assetResolver.ResolveCardAsset(cardID, dbfID)
	if image.Kind == "" {
		image.Kind = LiveCardStatsImageKindNone
		image.Value = ""
	}
	if image.Kind == LiveCardStatsImageKindNone {
		image.Value = ""
	}

	return LiveCardStatsCardPayload{
		Name:        name,
		Description: description,
		Image:       image,
	}
}

func (p *LiveCardStatsProjector) resolveCardInfo(current *card.Card) *carddb.CardInfo {
	if current == nil {
		return nil
	}

	switch info := current.Info.(type) {
	case *carddb.CardInfo:
		return info
	case carddb.CardInfo:
		resolved := info
		return &resolved
	}

	if p.cardDB == nil || current.ID == "" {
		return nil
	}

	info, ok := p.cardDB.GetCard(current.ID)
	if !ok {
		return nil
	}
	return info
}

func (p *LiveCardStatsProjector) buildPlayersSnapshot() LiveCardStatsPlayersSnapshot {
	selfCards := p.buildSortedCards(p.self)
	opponentCards := p.buildSortedCards(p.opponent)

	return LiveCardStatsPlayersSnapshot{
		Self: LiveCardStatsSideSnapshot{
			TotalPlays:  p.snapshot.Players.Self.TotalPlays,
			UniqueCards: len(selfCards),
			Cards:       selfCards,
		},
		Opponent: LiveCardStatsSideSnapshot{
			TotalPlays:  p.snapshot.Players.Opponent.TotalPlays,
			UniqueCards: len(opponentCards),
			Cards:       opponentCards,
		},
	}
}

func (p *LiveCardStatsProjector) buildSortedCards(source map[string]*liveCardStatsAggregate) []LiveCardStatsCardEntry {
	if len(source) == 0 {
		return make([]LiveCardStatsCardEntry, 0)
	}

	result := make([]LiveCardStatsCardEntry, 0, len(source))
	for _, aggregate := range source {
		if aggregate == nil {
			continue
		}
		result = append(result, cloneLiveCardStatsCardEntry(aggregate.entry))
	}

	slices.SortFunc(result, func(left, right LiveCardStatsCardEntry) int {
		switch {
		case left.FirstPlaySeq < right.FirstPlaySeq:
			return -1
		case left.FirstPlaySeq > right.FirstPlaySeq:
			return 1
		case left.CardID < right.CardID:
			return -1
		case left.CardID > right.CardID:
			return 1
		default:
			return 0
		}
	})

	return result
}

func (p *LiveCardStatsProjector) restoreAggregates(target map[string]*liveCardStatsAggregate, entries []LiveCardStatsCardEntry) {
	for _, entry := range entries {
		cloned := cloneLiveCardStatsCardEntry(entry)
		target[entry.CardID] = &liveCardStatsAggregate{entry: cloned}
	}
}

func (p *LiveCardStatsProjector) publishCurrentSnapshot() error {
	if p.sink == nil {
		return nil
	}
	if err := p.sink.Publish(p.Snapshot()); err != nil {
		return fmt.Errorf("publish live card stats snapshot: %w", err)
	}
	return nil
}

func marshalLiveCardStatsSnapshot(snapshot LiveCardStatsSnapshot) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(snapshot); err != nil {
		return nil, err
	}

	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
}

func normalizeLiveCardStatsSnapshot(snapshot LiveCardStatsSnapshot) (LiveCardStatsSnapshot, error) {
	if snapshot.SchemaVersion == "" {
		snapshot.SchemaVersion = LiveCardStatsSchemaVersion
	}
	if snapshot.SchemaVersion != LiveCardStatsSchemaVersion {
		return LiveCardStatsSnapshot{}, fmt.Errorf("unsupported live card stats schema version %q", snapshot.SchemaVersion)
	}
	if snapshot.Match.Status == "" {
		snapshot.Match.Status = LiveCardStatsMatchStatusIdle
	}
	if snapshot.RecoveryStatus == "" {
		snapshot.RecoveryStatus = RecoveryStatusFresh
	}
	if snapshot.Players.Self.Cards == nil {
		snapshot.Players.Self.Cards = make([]LiveCardStatsCardEntry, 0)
	}
	if snapshot.Players.Opponent.Cards == nil {
		snapshot.Players.Opponent.Cards = make([]LiveCardStatsCardEntry, 0)
	}
	snapshot.Players.Self.UniqueCards = len(snapshot.Players.Self.Cards)
	snapshot.Players.Opponent.UniqueCards = len(snapshot.Players.Opponent.Cards)

	return cloneLiveCardStatsSnapshot(snapshot), nil
}

func newIdleLiveCardStatsSnapshot() LiveCardStatsSnapshot {
	return LiveCardStatsSnapshot{
		SchemaVersion:     LiveCardStatsSchemaVersion,
		ProjectionVersion: 0,
		RecoveryStatus:    RecoveryStatusFresh,
		Match: LiveCardStatsMatchSnapshot{
			Status: LiveCardStatsMatchStatusIdle,
		},
		Players: newEmptyLiveCardStatsPlayersSnapshot(),
	}
}

func newEmptyLiveCardStatsPlayersSnapshot() LiveCardStatsPlayersSnapshot {
	return LiveCardStatsPlayersSnapshot{
		Self: LiveCardStatsSideSnapshot{
			Cards: make([]LiveCardStatsCardEntry, 0),
		},
		Opponent: LiveCardStatsSideSnapshot{
			Cards: make([]LiveCardStatsCardEntry, 0),
		},
	}
}

func buildLiveCardStatsMatchKey(gameID string) string {
	gameID = strings.TrimSpace(gameID)
	if gameID == "" {
		return ""
	}
	return "game:" + gameID
}

func cloneLiveCardStatsSnapshot(snapshot LiveCardStatsSnapshot) LiveCardStatsSnapshot {
	cloned := snapshot
	cloned.Players.Self.Cards = cloneLiveCardStatsCards(snapshot.Players.Self.Cards)
	cloned.Players.Opponent.Cards = cloneLiveCardStatsCards(snapshot.Players.Opponent.Cards)
	return cloned
}

func cloneLiveCardStatsCards(entries []LiveCardStatsCardEntry) []LiveCardStatsCardEntry {
	if entries == nil {
		return make([]LiveCardStatsCardEntry, 0)
	}

	cloned := make([]LiveCardStatsCardEntry, 0, len(entries))
	for _, entry := range entries {
		cloned = append(cloned, cloneLiveCardStatsCardEntry(entry))
	}
	return cloned
}

func cloneLiveCardStatsCardEntry(entry LiveCardStatsCardEntry) LiveCardStatsCardEntry {
	return LiveCardStatsCardEntry{
		CardID:       entry.CardID,
		PlayCount:    entry.PlayCount,
		FirstPlaySeq: entry.FirstPlaySeq,
		LastPlaySeq:  entry.LastPlaySeq,
		Card: LiveCardStatsCardPayload{
			Name:        entry.Card.Name,
			Description: entry.Card.Description,
			Image: LiveCardStatsImage{
				Kind:  entry.Card.Image.Kind,
				Value: entry.Card.Image.Value,
			},
		},
	}
}
