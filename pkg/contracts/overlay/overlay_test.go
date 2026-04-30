package overlay

import "testing"

func TestSnapshotValidateAcceptsOverlaySnapshotV1(t *testing.T) {
	t.Parallel()

	snapshot := NewSnapshot()
	snapshot.Match = Match{
		MatchID: "game:1",
		GameID:  "1",
		Status:  MatchStatusActive,
		DeckID:  "deck-1",
	}
	snapshot.PlayerDeckRemaining = []PlayerDeckCard{
		{
			CardID:          "CARD_1",
			Name:            "Карта 1",
			ManaCost:        2,
			Rarity:          "RARE",
			CardType:        "SPELL",
			InitialCopies:   2,
			RemainingCopies: 1,
			DrawnCopies:     1,
			PlayedCopies:    1,
		},
	}
	snapshot.OpponentPlayedCards = []OpponentPlayedCard{
		{
			CardID:        "CARD_2",
			Name:          "Карта 2",
			ManaCost:      3,
			PlayCount:     1,
			FirstPlayedAt: "2026-04-27T12:00:00Z",
			LastPlayedAt:  "2026-04-27T12:00:00Z",
		},
	}

	if err := snapshot.Validate(); err != nil {
		t.Fatalf("expected valid snapshot, got %v", err)
	}
}

func TestSnapshotValidateRejectsInvalidPayload(t *testing.T) {
	t.Parallel()

	snapshot := NewSnapshot()
	snapshot.SchemaVersion = "overlay_snapshot/v999"
	if err := snapshot.Validate(); err == nil {
		t.Fatalf("expected invalid schema version")
	}

	snapshot = NewSnapshot()
	snapshot.PlayerDeckRemaining = []PlayerDeckCard{
		{
			CardID:          "CARD_1",
			InitialCopies:   1,
			RemainingCopies: 2,
		},
	}
	if err := snapshot.Validate(); err == nil {
		t.Fatalf("expected invalid deck counts")
	}

	snapshot = NewSnapshot()
	snapshot.PlayerDeckRemaining = []PlayerDeckCard{
		{
			CardID:        "CARD_1",
			InitialCopies: 1,
			PlayedCopies:  2,
		},
	}
	if err := snapshot.Validate(); err == nil {
		t.Fatalf("expected invalid played counts")
	}
}

func TestEnvelopeValidateEnforcesMessagePayloadShape(t *testing.T) {
	t.Parallel()

	snapshot := NewSnapshot()
	snapshot.Match.Status = MatchStatusActive
	envelope := Envelope{
		Type:          MessageTypeSnapshot,
		SchemaVersion: SchemaVersionV1,
		Snapshot:      &snapshot,
	}
	if err := envelope.Validate(); err != nil {
		t.Fatalf("expected valid snapshot envelope, got %v", err)
	}

	invalid := Envelope{
		Type:          MessageTypeDelta,
		SchemaVersion: SchemaVersionV1,
	}
	if err := invalid.Validate(); err == nil {
		t.Fatalf("expected invalid delta envelope without payload")
	}
}

func TestNormalizeFillsDefaults(t *testing.T) {
	t.Parallel()

	snapshot := Snapshot{}
	snapshot.Normalize()

	if snapshot.SchemaVersion != SchemaVersionV1 {
		t.Fatalf("unexpected schema version: %q", snapshot.SchemaVersion)
	}
	if snapshot.Match.Status != MatchStatusIdle {
		t.Fatalf("unexpected match status: %q", snapshot.Match.Status)
	}
	if snapshot.PlayerDeckRemaining == nil || snapshot.OpponentPlayedCards == nil {
		t.Fatalf("expected empty slices after normalize")
	}
	if snapshot.Health.Monitor != HealthStateHealthy || snapshot.Health.Transport != HealthStateHealthy {
		t.Fatalf("unexpected normalized health: %#v", snapshot.Health)
	}
}
