package monitor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

type stubLiveCardStatsCardDB struct {
	cards map[string]*carddb.CardInfo
}

func (db *stubLiveCardStatsCardDB) Load(string) error {
	return nil
}

func (db *stubLiveCardStatsCardDB) GetCard(id string) (*carddb.CardInfo, bool) {
	info, ok := db.cards[id]
	return info, ok
}

func (db *stubLiveCardStatsCardDB) GetCardByDBFID(dbfID int) (*carddb.CardInfo, bool) {
	for _, info := range db.cards {
		if info != nil && info.DBFID == dbfID {
			return info, true
		}
	}
	return nil, false
}

func (db *stubLiveCardStatsCardDB) GetCardByName(string) (*carddb.CardInfo, bool) {
	return nil, false
}

type recordingLiveCardStatsSink struct {
	snapshots []LiveCardStatsSnapshot
}

func (s *recordingLiveCardStatsSink) Publish(snapshot LiveCardStatsSnapshot) error {
	s.snapshots = append(s.snapshots, snapshot)
	return nil
}

func TestLiveCardStatsAggregatesBySideAndCardID(t *testing.T) {
	t.Parallel()

	projector := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{
		CardDB: &stubLiveCardStatsCardDB{
			cards: map[string]*carddb.CardInfo{
				"CARD_A": {ID: "CARD_A", Name: "Карта А", Type: carddb.CardTypeSpell},
				"CARD_B": {ID: "CARD_B", Name: "Карта Б", Type: carddb.CardTypeMinion},
			},
		},
		Now: fixedTimeSequence(
			time.Unix(100, 0),
			time.Unix(101, 0),
			time.Unix(102, 0),
			time.Unix(103, 0),
			time.Unix(104, 0),
		),
	})

	mustProjectEvent(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "CARD_A", nil))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "CARD_A", nil))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypeOpponent, "CARD_A", nil))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypeOpponent, "CARD_B", nil))

	snapshot := projector.Snapshot()
	if snapshot.Players.Self.TotalPlays != 2 {
		t.Fatalf("unexpected self total plays: %d", snapshot.Players.Self.TotalPlays)
	}
	if snapshot.Players.Self.UniqueCards != 1 {
		t.Fatalf("unexpected self unique cards: %d", snapshot.Players.Self.UniqueCards)
	}
	if len(snapshot.Players.Self.Cards) != 1 {
		t.Fatalf("unexpected self cards: %#v", snapshot.Players.Self.Cards)
	}
	if snapshot.Players.Self.Cards[0].CardID != "CARD_A" || snapshot.Players.Self.Cards[0].PlayCount != 2 {
		t.Fatalf("unexpected self aggregate: %#v", snapshot.Players.Self.Cards[0])
	}

	if snapshot.Players.Opponent.TotalPlays != 2 {
		t.Fatalf("unexpected opponent total plays: %d", snapshot.Players.Opponent.TotalPlays)
	}
	if snapshot.Players.Opponent.UniqueCards != 2 {
		t.Fatalf("unexpected opponent unique cards: %d", snapshot.Players.Opponent.UniqueCards)
	}
	if len(snapshot.Players.Opponent.Cards) != 2 {
		t.Fatalf("unexpected opponent cards: %#v", snapshot.Players.Opponent.Cards)
	}
	if snapshot.Players.Opponent.Cards[0].CardID != "CARD_A" || snapshot.Players.Opponent.Cards[1].CardID != "CARD_B" {
		t.Fatalf("unexpected opponent aggregates: %#v", snapshot.Players.Opponent.Cards)
	}
}

func TestLiveCardStatsKeepsSelfAndOpponentInSeparateArrays(t *testing.T) {
	t.Parallel()

	projector := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{})

	mustProjectEvent(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "SELF_CARD", nil))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypeOpponent, "OPP_CARD", nil))

	snapshot := projector.Snapshot()
	if len(snapshot.Players.Self.Cards) != 1 || len(snapshot.Players.Opponent.Cards) != 1 {
		t.Fatalf("expected separate non-empty arrays, got self=%d opponent=%d", len(snapshot.Players.Self.Cards), len(snapshot.Players.Opponent.Cards))
	}
	if snapshot.Players.Self.Cards[0].CardID != "SELF_CARD" {
		t.Fatalf("unexpected self card array: %#v", snapshot.Players.Self.Cards)
	}
	if snapshot.Players.Opponent.Cards[0].CardID != "OPP_CARD" {
		t.Fatalf("unexpected opponent card array: %#v", snapshot.Players.Opponent.Cards)
	}
}

func TestLiveCardStatsIncrementsPlayCountForRepeatedCard(t *testing.T) {
	t.Parallel()

	projector := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{})

	mustProjectEvent(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "REPEAT_1", nil))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "REPEAT_1", nil))

	entry := projector.Snapshot().Players.Self.Cards[0]
	if entry.PlayCount != 2 {
		t.Fatalf("unexpected play count: %d", entry.PlayCount)
	}
	if entry.FirstPlaySeq != 1 || entry.LastPlaySeq != 2 {
		t.Fatalf("unexpected sequence bounds: %#v", entry)
	}
}

func TestLiveCardStatsUsesCardDBNameAndCleanDescription(t *testing.T) {
	t.Parallel()

	projector := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{
		CardDB: &stubLiveCardStatsCardDB{
			cards: map[string]*carddb.CardInfo{
				"CLEAN_1": {
					ID:   "CLEAN_1",
					Name: "Очищенная карта",
					Text: "[x]Наносит <b>6</b> ед. урона.<br>Еще строка.",
					Type: carddb.CardTypeSpell,
				},
			},
		},
	})

	mustProjectEvent(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "CLEAN_1", nil))

	entry := projector.Snapshot().Players.Self.Cards[0]
	if entry.Card.Name != "Очищенная карта" {
		t.Fatalf("unexpected card name: %q", entry.Card.Name)
	}
	if entry.Card.Description != "Наносит 6 ед. урона.\nЕще строка." {
		t.Fatalf("unexpected card description: %q", entry.Card.Description)
	}
	if entry.Card.Image.Kind != LiveCardStatsImageKindNone {
		t.Fatalf("unexpected default image payload: %#v", entry.Card.Image)
	}
}

func TestLiveCardStatsFallsBackWhenCardMetadataMissing(t *testing.T) {
	t.Parallel()

	projector := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{})

	mustProjectEvent(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "UNKNOWN_CARD", nil))

	entry := projector.Snapshot().Players.Self.Cards[0]
	if entry.Card.Name != "UNKNOWN_CARD" {
		t.Fatalf("unexpected fallback name: %q", entry.Card.Name)
	}
	if entry.Card.Description != "" {
		t.Fatalf("unexpected fallback description: %q", entry.Card.Description)
	}
	if entry.Card.Image.Kind != LiveCardStatsImageKindNone {
		t.Fatalf("unexpected fallback image: %#v", entry.Card.Image)
	}
}

func TestLiveCardStatsIgnoresEventsWithoutPlayableCardID(t *testing.T) {
	t.Parallel()

	projector := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{
		CardDB: &stubLiveCardStatsCardDB{
			cards: map[string]*carddb.CardInfo{
				"HERO_POWER": {ID: "HERO_POWER", Name: "Сила героя", Type: carddb.CardTypeHeroPower},
			},
		},
	})

	mustProjectEvent(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustProjectEvent(t, projector, &game.CardPlayedEvent{Player: types.PlayerTypePlayer})
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "", nil))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "HERO_POWER", &carddb.CardInfo{ID: "HERO_POWER", Type: carddb.CardTypeHeroPower}))

	snapshot := projector.Snapshot()
	if snapshot.Players.Self.TotalPlays != 0 || snapshot.Players.Self.UniqueCards != 0 || len(snapshot.Players.Self.Cards) != 0 {
		t.Fatalf("expected ignored events to produce no stats: %#v", snapshot.Players.Self)
	}
}

func TestLiveCardStatsResetsOnGameStart(t *testing.T) {
	t.Parallel()

	projector := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{})

	mustProjectEvent(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "CARD_1", nil))
	mustProjectEvent(t, projector, game.NewGameStartEvent("game-2", state.GameModeStandard))

	snapshot := projector.Snapshot()
	if snapshot.Match.Key != "game:game-2" {
		t.Fatalf("unexpected match key after reset: %q", snapshot.Match.Key)
	}
	if snapshot.Match.Status != LiveCardStatsMatchStatusActive {
		t.Fatalf("unexpected match status after reset: %q", snapshot.Match.Status)
	}
	if snapshot.RecoveryStatus != RecoveryStatusFresh {
		t.Fatalf("unexpected recovery status after reset: %q", snapshot.RecoveryStatus)
	}
	if snapshot.Players.Self.TotalPlays != 0 || len(snapshot.Players.Self.Cards) != 0 {
		t.Fatalf("expected self stats reset, got %#v", snapshot.Players.Self)
	}
}

func TestLiveCardStatsFinalizesOnGameEndWithoutDataLoss(t *testing.T) {
	t.Parallel()

	projector := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{})

	mustProjectEvent(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "CARD_1", nil))
	mustProjectEvent(t, projector, game.NewGameEndEvent(types.PlayerTypePlayer))

	snapshot := projector.Snapshot()
	if snapshot.Match.Status != LiveCardStatsMatchStatusFinalized {
		t.Fatalf("unexpected finalized status: %q", snapshot.Match.Status)
	}
	if snapshot.Players.Self.TotalPlays != 1 || len(snapshot.Players.Self.Cards) != 1 {
		t.Fatalf("expected finalized snapshot to keep card data, got %#v", snapshot.Players.Self)
	}
}

func TestLiveCardStatsDoesNotLeakHiddenOpponentInformation(t *testing.T) {
	t.Parallel()

	projector := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{})

	mustProjectEvent(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	if err := projector.Process(game.NewDeckListEvent(types.PlayerTypeOpponent, []*card.Card{card.New("HIDDEN_1", 1, types.PlayerTypeOpponent)}, "", "", "")); err != nil {
		t.Fatalf("process deck list: %v", err)
	}
	if err := projector.Process(game.NewCardDrawnEvent(types.PlayerTypeOpponent, card.New("HIDDEN_2", 2, types.PlayerTypeOpponent))); err != nil {
		t.Fatalf("process card drawn: %v", err)
	}
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypeOpponent, "VISIBLE_1", nil))

	snapshot := projector.Snapshot()
	if snapshot.Players.Opponent.TotalPlays != 1 || len(snapshot.Players.Opponent.Cards) != 1 {
		t.Fatalf("unexpected opponent public projection: %#v", snapshot.Players.Opponent)
	}
	if snapshot.Players.Opponent.Cards[0].CardID != "VISIBLE_1" {
		t.Fatalf("unexpected visible opponent card: %#v", snapshot.Players.Opponent.Cards)
	}
}

func TestLiveCardStatsSnapshotFileAtomicReplace(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "live_card_stats.json")
	sink, err := NewAtomicFileLiveCardStatsSink(path)
	if err != nil {
		t.Fatalf("create sink: %v", err)
	}

	first := LiveCardStatsSnapshot{
		SchemaVersion:     LiveCardStatsSchemaVersion,
		ProjectionVersion: 1,
		RecoveryStatus:    RecoveryStatusFresh,
		Match: LiveCardStatsMatchSnapshot{
			Key:    "game:1",
			Status: LiveCardStatsMatchStatusActive,
		},
		Players: LiveCardStatsPlayersSnapshot{
			Self: LiveCardStatsSideSnapshot{
				TotalPlays:  1,
				UniqueCards: 1,
				Cards: []LiveCardStatsCardEntry{
					{CardID: "CARD_1", PlayCount: 1, FirstPlaySeq: 1, LastPlaySeq: 1},
				},
			},
			Opponent: LiveCardStatsSideSnapshot{Cards: make([]LiveCardStatsCardEntry, 0)},
		},
	}
	second := first
	second.ProjectionVersion = 2
	second.Match.Status = LiveCardStatsMatchStatusFinalized

	if err := sink.Publish(first); err != nil {
		t.Fatalf("publish first snapshot: %v", err)
	}
	if err := sink.Publish(second); err != nil {
		t.Fatalf("publish second snapshot: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot file: %v", err)
	}
	if strings.Contains(string(data), `"projection_version": 1`) {
		t.Fatalf("expected replaced file content, got:\n%s", string(data))
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("expected no temporary file left behind, got err=%v", err)
	}
}

func TestLiveCardStatsRecoveryStatusAfterRestart(t *testing.T) {
	t.Parallel()

	original := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{
		CardDB: &stubLiveCardStatsCardDB{
			cards: map[string]*carddb.CardInfo{
				"RESUME_1": {ID: "RESUME_1", Name: "Возврат", Type: carddb.CardTypeSpell},
			},
		},
	})

	mustProjectEvent(t, original, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustProjectEvent(t, original, newPlayedCardEvent(types.PlayerTypePlayer, "RESUME_1", nil))

	persisted := original.PersistenceSnapshot()
	if persisted == nil {
		t.Fatalf("expected persistence snapshot")
	}

	sink := &recordingLiveCardStatsSink{}
	resumed := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{
		CardDB: &stubLiveCardStatsCardDB{
			cards: map[string]*carddb.CardInfo{
				"RESUME_1": {ID: "RESUME_1", Name: "Возврат", Type: carddb.CardTypeSpell},
			},
		},
		Sink: sink,
	})

	if err := resumed.Restore(*persisted); err != nil {
		t.Fatalf("restore snapshot: %v", err)
	}

	snapshot := resumed.Snapshot()
	if snapshot.RecoveryStatus != RecoveryStatusResumed {
		t.Fatalf("unexpected recovery status after restore: %q", snapshot.RecoveryStatus)
	}
	if len(sink.snapshots) != 1 {
		t.Fatalf("expected restore to republish one snapshot, got %d", len(sink.snapshots))
	}

	mustProjectEvent(t, resumed, newPlayedCardEvent(types.PlayerTypePlayer, "RESUME_1", nil))
	entry := resumed.Snapshot().Players.Self.Cards[0]
	if entry.PlayCount != 2 || entry.FirstPlaySeq != 1 || entry.LastPlaySeq != 2 {
		t.Fatalf("unexpected resumed entry: %#v", entry)
	}
}

func TestLiveCardStatsSnapshotJSONContract(t *testing.T) {
	t.Parallel()

	projector := NewLiveCardStatsProjector(LiveCardStatsProjectorOptions{})
	mustProjectEvent(t, projector, game.NewGameStartEvent("game-1", state.GameModeStandard))
	mustProjectEvent(t, projector, newPlayedCardEvent(types.PlayerTypePlayer, "CARD_1", nil))

	payload, err := marshalLiveCardStatsSnapshot(projector.Snapshot())
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}

	var decoded struct {
		SchemaVersion string `json:"schema_version"`
		Players       struct {
			Self struct {
				Cards []json.RawMessage `json:"cards"`
			} `json:"self"`
			Opponent struct {
				Cards []json.RawMessage `json:"cards"`
			} `json:"opponent"`
		} `json:"players"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if decoded.SchemaVersion != LiveCardStatsSchemaVersion {
		t.Fatalf("unexpected schema version: %q", decoded.SchemaVersion)
	}
	if len(decoded.Players.Self.Cards) != 1 || len(decoded.Players.Opponent.Cards) != 0 {
		t.Fatalf("unexpected separate arrays in contract: %#v", decoded.Players)
	}
}

func mustProjectEvent(t *testing.T, projector *LiveCardStatsProjector, event game.Event) {
	t.Helper()
	if err := projector.Process(event); err != nil {
		t.Fatalf("process event %T: %v", event, err)
	}
}

func newPlayedCardEvent(player types.PlayerType, cardID string, info *carddb.CardInfo) *game.CardPlayedEvent {
	played := card.New(cardID, 1, player)
	if info != nil {
		played.Info = info
	}
	return game.NewCardPlayedEvent(player, played, 0)
}

func fixedTimeSequence(times ...time.Time) func() time.Time {
	if len(times) == 0 {
		return time.Now
	}

	index := 0
	return func() time.Time {
		if index >= len(times) {
			return times[len(times)-1]
		}
		current := times[index]
		index++
		return current
	}
}
