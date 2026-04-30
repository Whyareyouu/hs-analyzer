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
	"hsanalyzer/pkg/contracts/overlay"
)

type OverlayCardAssetResolver interface {
	ResolveCardImage(cardID string, dbfID int) string
}

type OverlaySnapshotSink interface {
	Publish(snapshot overlay.Snapshot) error
}

type OverlaySnapshotProjectorOptions struct {
	CardDB        carddb.CardDatabase
	AssetResolver OverlayCardAssetResolver
	Sink          OverlaySnapshotSink
	Now           func() time.Time
}

type OverlaySnapshotProjector struct {
	cardDB        carddb.CardDatabase
	assetResolver OverlayCardAssetResolver
	sink          OverlaySnapshotSink
	now           func() time.Time

	snapshot          overlay.Snapshot
	templateDeck      []overlay.PlayerDeckCard
	templateDeckID    string
	templateName      string
	activeDeck        map[string]*overlay.PlayerDeckCard
	selfPlayed        map[string]int
	opponentCards     map[string]*overlay.OpponentPlayedCard
	preGameCardEvents []game.Event
}

type NoopOverlayCardAssetResolver struct{}

func (NoopOverlayCardAssetResolver) ResolveCardImage(string, int) string {
	return ""
}

type AtomicFileOverlaySnapshotSink struct {
	path        string
	lastPayload []byte
}

func NewAtomicFileOverlaySnapshotSink(path string) (*AtomicFileOverlaySnapshotSink, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("overlay snapshot path is empty")
	}
	return &AtomicFileOverlaySnapshotSink{path: filepath.Clean(path)}, nil
}

func (s *AtomicFileOverlaySnapshotSink) Publish(snapshot overlay.Snapshot) error {
	if s == nil {
		return errors.New("overlay snapshot sink is nil")
	}

	payload, err := marshalOverlaySnapshot(snapshot)
	if err != nil {
		return fmt.Errorf("marshal overlay snapshot: %w", err)
	}
	if bytes.Equal(payload, s.lastPayload) {
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
			return fmt.Errorf("read existing overlay snapshot: %w", err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create overlay snapshot directory: %w", err)
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, payload, 0o644); err != nil {
		return fmt.Errorf("write temporary overlay snapshot: %w", err)
	}
	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("replace overlay snapshot: %w", err)
	}

	s.lastPayload = append([]byte(nil), payload...)
	return nil
}

func NewOverlaySnapshotProjector(options OverlaySnapshotProjectorOptions) *OverlaySnapshotProjector {
	now := options.Now
	if now == nil {
		now = time.Now
	}

	resolver := options.AssetResolver
	if resolver == nil {
		resolver = NoopOverlayCardAssetResolver{}
	}

	snapshot := overlay.NewSnapshot()
	return &OverlaySnapshotProjector{
		cardDB:            options.CardDB,
		assetResolver:     resolver,
		sink:              options.Sink,
		now:               now,
		snapshot:          snapshot,
		templateDeck:      make([]overlay.PlayerDeckCard, 0),
		activeDeck:        make(map[string]*overlay.PlayerDeckCard),
		selfPlayed:        make(map[string]int),
		opponentCards:     make(map[string]*overlay.OpponentPlayedCard),
		preGameCardEvents: make([]game.Event, 0),
	}
}

func (p *OverlaySnapshotProjector) Process(event game.Event) error {
	if p == nil || event == nil {
		return nil
	}

	switch current := event.(type) {
	case *game.DeckListEvent:
		return p.processDeckList(current)
	case *game.GameStartEvent:
		return p.processGameStart(current)
	case *game.CardDrawnEvent:
		if p.snapshot.Match.Status != overlay.MatchStatusActive {
			p.rememberPreGameCardEvent(current)
			return nil
		}
		return p.processCardDrawn(current)
	case *game.CardPlayedEvent:
		if p.snapshot.Match.Status != overlay.MatchStatusActive {
			p.rememberPreGameCardEvent(current)
			return nil
		}
		return p.processCardPlayed(current)
	case *game.GameEndEvent:
		return p.processGameEnd(current)
	default:
		return nil
	}
}

func (p *OverlaySnapshotProjector) Restore(snapshot overlay.Snapshot) error {
	if p == nil {
		return errors.New("overlay snapshot projector is nil")
	}
	snapshot.Normalize()
	if err := snapshot.Validate(); err != nil {
		return err
	}

	p.snapshot = snapshot.Clone()
	p.templateDeck = clonePlayerDeckCards(snapshot.PlayerDeckRemaining)
	p.templateDeckID = snapshot.Match.DeckID
	p.templateName = snapshot.Match.DeckName
	p.activeDeck = make(map[string]*overlay.PlayerDeckCard, len(snapshot.PlayerDeckRemaining))
	p.selfPlayed = make(map[string]int, len(snapshot.PlayerDeckRemaining))
	p.preGameCardEvents = p.preGameCardEvents[:0]
	for _, current := range snapshot.PlayerDeckRemaining {
		card := current
		p.activeDeck[current.CardID] = &card
		switch {
		case current.PlayedCopies > 0:
			p.selfPlayed[current.CardID] = current.PlayedCopies
		default:
			outOfDeckCopies := current.InitialCopies - current.RemainingCopies
			if outOfDeckCopies > 0 {
				p.selfPlayed[current.CardID] = minInt(outOfDeckCopies, current.DrawnCopies)
			}
		}
	}
	p.opponentCards = make(map[string]*overlay.OpponentPlayedCard, len(snapshot.OpponentPlayedCards))
	for _, current := range snapshot.OpponentPlayedCards {
		card := current
		p.opponentCards[current.CardID] = &card
	}
	return p.publish()
}

func (p *OverlaySnapshotProjector) Snapshot() overlay.Snapshot {
	if p == nil {
		return overlay.NewSnapshot()
	}
	return p.snapshot.Clone()
}

func (p *OverlaySnapshotProjector) PersistenceSnapshot() *overlay.Snapshot {
	if p == nil {
		return nil
	}
	snapshot := p.Snapshot()
	return &snapshot
}

func (p *OverlaySnapshotProjector) SetHealth(health overlay.Health) error {
	if p == nil {
		return nil
	}

	normalized := overlay.NewSnapshot().Health
	if strings.TrimSpace(health.Monitor) != "" {
		normalized.Monitor = health.Monitor
	}
	if strings.TrimSpace(health.Transport) != "" {
		normalized.Transport = health.Transport
	}
	if p.snapshot.Health == normalized {
		return nil
	}

	p.snapshot.Health = normalized
	return p.publish()
}

func (p *OverlaySnapshotProjector) processDeckList(event *game.DeckListEvent) error {
	if event == nil || event.Player != types.PlayerTypePlayer {
		return nil
	}

	grouped := make(map[string]*overlay.PlayerDeckCard)
	for _, deckCard := range event.Cards {
		if deckCard == nil || strings.TrimSpace(deckCard.ID) == "" {
			continue
		}
		info := p.resolveCardInfo(deckCard)
		entry, ok := grouped[deckCard.ID]
		if !ok {
			cardModel := p.buildPlayerDeckCard(deckCard, info)
			entry = &cardModel
			grouped[deckCard.ID] = entry
		}
		entry.InitialCopies++
		entry.RemainingCopies++
	}

	cards := make([]overlay.PlayerDeckCard, 0, len(grouped))
	for _, current := range grouped {
		cards = append(cards, *current)
	}
	sortPlayerDeckCards(cards)

	p.templateDeck = clonePlayerDeckCards(cards)
	p.templateDeckID = strings.TrimSpace(event.DeckID)
	p.templateName = strings.TrimSpace(event.Name)
	if p.templateName == "" {
		p.templateName = p.templateDeckID
	}

	p.snapshot.Match.DeckID = p.templateDeckID
	p.snapshot.Match.DeckName = p.templateName
	p.resetActiveDeckFromTemplate()
	p.preGameCardEvents = p.preGameCardEvents[:0]
	return p.publish()
}

func (p *OverlaySnapshotProjector) processGameStart(event *game.GameStartEvent) error {
	if event == nil {
		return nil
	}

	p.snapshot = overlay.NewSnapshot()
	p.snapshot.Match = overlay.Match{
		MatchID:  buildLiveCardStatsMatchKey(event.GameID),
		GameID:   strings.TrimSpace(event.GameID),
		Status:   overlay.MatchStatusActive,
		DeckID:   p.templateDeckID,
		DeckName: p.templateName,
	}
	p.resetActiveDeckFromTemplate()
	p.opponentCards = make(map[string]*overlay.OpponentPlayedCard)
	for _, pending := range p.drainPreGameCardEvents() {
		switch current := pending.(type) {
		case *game.CardDrawnEvent:
			p.applyCardDrawn(current)
		case *game.CardPlayedEvent:
			p.applyCardPlayed(current)
		}
	}
	return p.publish()
}

func (p *OverlaySnapshotProjector) processCardDrawn(event *game.CardDrawnEvent) error {
	if !p.applyCardDrawn(event) {
		return nil
	}
	return p.publish()
}

func (p *OverlaySnapshotProjector) processCardPlayed(event *game.CardPlayedEvent) error {
	if !p.applyCardPlayed(event) {
		return nil
	}
	return p.publish()
}

func (p *OverlaySnapshotProjector) applyCardDrawn(event *game.CardDrawnEvent) bool {
	if event == nil || event.Card == nil {
		return false
	}
	if p.snapshot.Match.Status != overlay.MatchStatusActive {
		return false
	}

	cardID := strings.TrimSpace(event.Card.ID)
	if cardID == "" {
		return false
	}

	current, ok := p.activeDeck[cardID]
	if !ok || current == nil || current.RemainingCopies == 0 {
		return false
	}

	// Early Zone.log draw events can arrive before side correction has enough
	// evidence, but a resolved selected-deck card still belongs in our deck.
	current.RemainingCopies--
	current.DrawnCopies++
	p.rebuildPlayerDeckSnapshot()
	return true
}

func (p *OverlaySnapshotProjector) applyCardPlayed(event *game.CardPlayedEvent) bool {
	if event == nil || event.Card == nil {
		return false
	}
	if p.snapshot.Match.Status != overlay.MatchStatusActive {
		return false
	}

	cardID := strings.TrimSpace(event.Card.ID)
	if cardID == "" {
		return false
	}

	if event.Player == types.PlayerTypePlayer {
		current, ok := p.activeDeck[cardID]
		if !ok || current == nil {
			return false
		}

		playedCopies := p.selfPlayed[cardID]
		if playedCopies >= current.DrawnCopies && current.RemainingCopies > 0 {
			current.RemainingCopies--
			current.DrawnCopies++
		}
		p.selfPlayed[cardID] = playedCopies + 1
		current.PlayedCopies = minInt(p.selfPlayed[cardID], current.InitialCopies)
		p.rebuildPlayerDeckSnapshot()
		return true
	}
	if event.Player != types.PlayerTypeOpponent {
		return false
	}

	current, ok := p.opponentCards[cardID]
	if !ok {
		entry := p.buildOpponentPlayedCard(event.Card, p.resolveCardInfo(event.Card))
		current = &entry
		p.opponentCards[cardID] = current
	}

	now := p.now().UTC().Format(time.RFC3339Nano)
	current.PlayCount++
	if current.FirstPlayedAt == "" {
		current.FirstPlayedAt = now
	}
	current.LastPlayedAt = now
	p.rebuildOpponentSnapshot()
	return true
}

func (p *OverlaySnapshotProjector) processGameEnd(*game.GameEndEvent) error {
	if p.snapshot.Match.Status != overlay.MatchStatusActive {
		return nil
	}
	p.snapshot.Match.Status = overlay.MatchStatusFinalized
	return p.publish()
}

func (p *OverlaySnapshotProjector) rememberPreGameCardEvent(event game.Event) {
	if event == nil {
		return
	}
	const maxPreGameCardEvents = 64
	if len(p.preGameCardEvents) >= maxPreGameCardEvents {
		copy(p.preGameCardEvents, p.preGameCardEvents[1:])
		p.preGameCardEvents[len(p.preGameCardEvents)-1] = event
		return
	}
	p.preGameCardEvents = append(p.preGameCardEvents, event)
}

func (p *OverlaySnapshotProjector) drainPreGameCardEvents() []game.Event {
	if len(p.preGameCardEvents) == 0 {
		return nil
	}
	events := append([]game.Event(nil), p.preGameCardEvents...)
	p.preGameCardEvents = p.preGameCardEvents[:0]
	return events
}

func (p *OverlaySnapshotProjector) buildPlayerDeckCard(current *card.Card, info *carddb.CardInfo) overlay.PlayerDeckCard {
	name, description, image, manaCost := p.resolveCardPresentation(current, info)
	rarity := ""
	cardType := ""
	isLegendary := false
	if info != nil {
		rarity = string(info.Rarity)
		cardType = string(info.Type)
		isLegendary = info.Rarity == carddb.RarityLegendary
	}

	return overlay.PlayerDeckCard{
		CardID:          strings.TrimSpace(current.ID),
		Name:            name,
		Description:     description,
		Image:           image,
		ManaCost:        manaCost,
		Rarity:          rarity,
		CardType:        cardType,
		IsLegendary:     isLegendary,
		InitialCopies:   0,
		RemainingCopies: 0,
		DrawnCopies:     0,
		PlayedCopies:    0,
	}
}

func (p *OverlaySnapshotProjector) buildOpponentPlayedCard(current *card.Card, info *carddb.CardInfo) overlay.OpponentPlayedCard {
	name, description, image, manaCost := p.resolveCardPresentation(current, info)
	return overlay.OpponentPlayedCard{
		CardID:      strings.TrimSpace(current.ID),
		Name:        name,
		Description: description,
		Image:       image,
		ManaCost:    manaCost,
	}
}

func (p *OverlaySnapshotProjector) resolveCardPresentation(current *card.Card, info *carddb.CardInfo) (string, string, string, int) {
	cardID := ""
	manaCost := 0
	if current != nil {
		cardID = strings.TrimSpace(current.ID)
		manaCost = current.Cost
	}
	if info == nil {
		info = p.resolveCardInfo(current)
	}

	name := cardID
	description := ""
	dbfID := 0
	if info != nil {
		if strings.TrimSpace(info.Name) != "" {
			name = info.Name
		}
		description = textutil.CleanCardText(info.Text)
		dbfID = info.DBFID
		if info.Cost > 0 || manaCost == 0 {
			manaCost = info.Cost
		}
	}

	image := p.assetResolver.ResolveCardImage(cardID, dbfID)
	return name, description, image, manaCost
}

func (p *OverlaySnapshotProjector) resolveCardInfo(current *card.Card) *carddb.CardInfo {
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

	if p.cardDB == nil || strings.TrimSpace(current.ID) == "" {
		return nil
	}

	info, ok := p.cardDB.GetCard(current.ID)
	if !ok {
		return nil
	}
	return info
}

func (p *OverlaySnapshotProjector) resetActiveDeckFromTemplate() {
	p.activeDeck = make(map[string]*overlay.PlayerDeckCard, len(p.templateDeck))
	p.selfPlayed = make(map[string]int, len(p.templateDeck))
	for _, current := range p.templateDeck {
		card := current
		card.RemainingCopies = card.InitialCopies
		card.DrawnCopies = 0
		card.PlayedCopies = 0
		p.activeDeck[current.CardID] = &card
	}
	p.rebuildPlayerDeckSnapshot()
}

func (p *OverlaySnapshotProjector) rebuildPlayerDeckSnapshot() {
	cards := make([]overlay.PlayerDeckCard, 0, len(p.activeDeck))
	for _, current := range p.activeDeck {
		if current == nil {
			continue
		}
		cards = append(cards, *current)
	}
	sortPlayerDeckCards(cards)
	p.snapshot.PlayerDeckRemaining = cards
}

func (p *OverlaySnapshotProjector) rebuildOpponentSnapshot() {
	cards := make([]overlay.OpponentPlayedCard, 0, len(p.opponentCards))
	for _, current := range p.opponentCards {
		if current == nil {
			continue
		}
		cards = append(cards, *current)
	}
	slices.SortFunc(cards, func(left, right overlay.OpponentPlayedCard) int {
		switch {
		case left.FirstPlayedAt < right.FirstPlayedAt:
			return -1
		case left.FirstPlayedAt > right.FirstPlayedAt:
			return 1
		case left.LastPlayedAt < right.LastPlayedAt:
			return -1
		case left.LastPlayedAt > right.LastPlayedAt:
			return 1
		case left.CardID < right.CardID:
			return -1
		case left.CardID > right.CardID:
			return 1
		default:
			return 0
		}
	})
	p.snapshot.OpponentPlayedCards = cards
}

func (p *OverlaySnapshotProjector) publish() error {
	p.snapshot.Normalize()
	if err := p.snapshot.Validate(); err != nil {
		return err
	}
	if p.sink == nil {
		return nil
	}
	return p.sink.Publish(p.snapshot.Clone())
}

func marshalOverlaySnapshot(snapshot overlay.Snapshot) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
}

func sortPlayerDeckCards(cards []overlay.PlayerDeckCard) {
	slices.SortFunc(cards, func(left, right overlay.PlayerDeckCard) int {
		switch {
		case left.ManaCost < right.ManaCost:
			return -1
		case left.ManaCost > right.ManaCost:
			return 1
		case left.Name < right.Name:
			return -1
		case left.Name > right.Name:
			return 1
		case left.CardID < right.CardID:
			return -1
		case left.CardID > right.CardID:
			return 1
		default:
			return 0
		}
	})
}

func clonePlayerDeckCards(cards []overlay.PlayerDeckCard) []overlay.PlayerDeckCard {
	return append([]overlay.PlayerDeckCard(nil), cards...)
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
