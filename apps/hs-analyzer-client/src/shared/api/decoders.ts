import type { DeckCard, DeckHistoryEntry, DeckListItem, DeckOverview, MatchupEntry } from "@/shared/model/dashboard";
import type { OverlayEnvelope, OverlaySnapshot } from "@/shared/model/overlay";

type UnknownRecord = Record<string, unknown>;

function ensureRecord(value: unknown, label: string): UnknownRecord {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value as UnknownRecord;
}

function ensureString(value: unknown, label: string): string {
  if (typeof value !== "string") {
    throw new Error(`${label} must be a string`);
  }
  return value;
}

function ensureNumber(value: unknown, label: string): number {
  if (typeof value !== "number" || Number.isNaN(value)) {
    throw new Error(`${label} must be a number`);
  }
  return value;
}

function optionalNumber(value: unknown): number {
  return typeof value === "number" && !Number.isNaN(value) ? value : 0;
}

function ensureArray(value: unknown, label: string): unknown[] {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value;
}

function optionalString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function parseMatchupEntry(value: unknown): MatchupEntry {
  const record = ensureRecord(value, "matchup entry");
  return {
    opponentClass: optionalString(record.opponent_class),
    opponentArchetype: optionalString(record.opponent_archetype),
    wins: ensureNumber(record.wins, "matchup wins"),
    losses: ensureNumber(record.losses, "matchup losses"),
    winrate: ensureNumber(record.winrate, "matchup winrate"),
    matches: ensureNumber(record.matches, "matchup matches"),
  };
}

export function parseOverlaySnapshot(value: unknown): OverlaySnapshot {
  const record = ensureRecord(value, "overlay snapshot");
  const match = ensureRecord(record.match, "overlay match");
  const health = ensureRecord(record.health, "overlay health");

  return {
    schemaVersion: ensureString(record.schema_version, "overlay schema_version"),
    match: {
      matchId: optionalString(match.match_id),
      gameId: optionalString(match.game_id),
      status: ensureString(match.status, "overlay match status") as OverlaySnapshot["match"]["status"],
      playerClass: optionalString(match.player_class),
      opponentClass: optionalString(match.opponent_class),
      deckId: optionalString(match.deck_id),
      deckName: optionalString(match.deck_name),
    },
    playerDeckRemaining: ensureArray(record.player_deck_remaining, "player deck remaining").map((item) => {
      const current = ensureRecord(item, "player deck card");
      return {
        cardId: ensureString(current.card_id, "player deck card id"),
        name: ensureString(current.name, "player deck card name"),
        description: optionalString(current.description),
        image: optionalString(current.image),
        manaCost: ensureNumber(current.mana_cost, "player deck card mana cost"),
        rarity: optionalString(current.rarity),
        cardType: optionalString(current.card_type),
        isLegendary: Boolean(current.is_legendary),
        initialCopies: ensureNumber(current.initial_copies, "player deck initial copies"),
        remainingCopies: ensureNumber(current.remaining_copies, "player deck remaining copies"),
        drawnCopies: ensureNumber(current.drawn_copies, "player deck drawn copies"),
        playedCopies: optionalNumber(current.played_copies),
      };
    }),
    playerOtherPlayedCards: ensureArray(record.player_other_played_cards ?? [], "player other played cards").map((item) => {
      const current = ensureRecord(item, "player other played card");
      return {
        cardId: ensureString(current.card_id, "player other played card id"),
        name: ensureString(current.name, "player other played card name"),
        description: optionalString(current.description),
        image: optionalString(current.image),
        manaCost: ensureNumber(current.mana_cost, "player other played card mana cost"),
        playCount: ensureNumber(current.play_count, "player other played card count"),
        firstPlayedAt: optionalString(current.first_played_at),
        lastPlayedAt: optionalString(current.last_played_at),
      };
    }),
    opponentPlayedCards: ensureArray(record.opponent_played_cards, "opponent played cards").map((item) => {
      const current = ensureRecord(item, "opponent played card");
      return {
        cardId: ensureString(current.card_id, "opponent played card id"),
        name: ensureString(current.name, "opponent played card name"),
        description: optionalString(current.description),
        image: optionalString(current.image),
        manaCost: ensureNumber(current.mana_cost, "opponent played card mana cost"),
        playCount: ensureNumber(current.play_count, "opponent played card count"),
        firstPlayedAt: optionalString(current.first_played_at),
        lastPlayedAt: optionalString(current.last_played_at),
      };
    }),
    health: {
      monitor: ensureString(health.monitor, "overlay monitor health") as OverlaySnapshot["health"]["monitor"],
      transport: ensureString(health.transport, "overlay transport health") as OverlaySnapshot["health"]["transport"],
    },
  };
}

export function parseOverlayEnvelope(value: unknown): OverlayEnvelope {
  const record = ensureRecord(value, "overlay envelope");
  return {
    type: ensureString(record.type, "overlay envelope type") as OverlayEnvelope["type"],
    schemaVersion: ensureString(record.schema_version, "overlay envelope schema_version"),
    snapshot: record.snapshot ? parseOverlaySnapshot(record.snapshot) : undefined,
    delta: record.delta ? parseOverlaySnapshot(record.delta) : undefined,
    health: record.health
      ? parseOverlaySnapshot({
          schema_version: "overlay_snapshot/v1",
          match: { status: "idle" },
          player_deck_remaining: [],
          player_other_played_cards: [],
          opponent_played_cards: [],
          health: record.health,
        }).health
      : undefined,
    matchId: optionalString(record.match_id),
  };
}

export function parseDeckList(value: unknown): DeckListItem[] {
  return ensureArray(value, "deck list").map((item) => {
    const record = ensureRecord(item, "deck list item");
    return {
      deckId: ensureString(record.deck_id, "deck id"),
      name: ensureString(record.name, "deck name"),
      className: ensureString(record.class, "deck class"),
      wins: ensureNumber(record.wins, "deck wins"),
      losses: ensureNumber(record.losses, "deck losses"),
      winrate: ensureNumber(record.winrate, "deck winrate"),
      lastPlayedAt: optionalString(record.last_played_at),
      icon: optionalString(record.icon),
    };
  });
}

export function parseDeckOverview(value: unknown): DeckOverview {
  const record = ensureRecord(value, "deck overview");
  return {
    deckId: ensureString(record.deck_id, "overview deck id"),
    name: ensureString(record.name, "overview deck name"),
    className: ensureString(record.class, "overview class"),
    wins: ensureNumber(record.wins, "overview wins"),
    losses: ensureNumber(record.losses, "overview losses"),
    totalMatches: ensureNumber(record.total_matches, "overview total matches"),
    winrate: ensureNumber(record.winrate, "overview winrate"),
    bestMatchups: ensureArray(record.best_matchups, "best matchups").map(parseMatchupEntry),
    worstMatchups: ensureArray(record.worst_matchups, "worst matchups").map(parseMatchupEntry),
  };
}

export function parseDeckCards(value: unknown): DeckCard[] {
  return ensureArray(value, "deck cards").map((item) => {
    const record = ensureRecord(item, "deck card");
    return {
      cardId: ensureString(record.card_id, "deck card id"),
      name: ensureString(record.name, "deck card name"),
      manaCost: ensureNumber(record.mana_cost, "deck card mana cost"),
      copies: ensureNumber(record.copies, "deck card copies"),
      rarity: optionalString(record.rarity),
      isLegendary: Boolean(record.is_legendary),
      image: optionalString(record.image),
    };
  });
}

export function parseMatchups(value: unknown): MatchupEntry[] {
  return ensureArray(value, "matchups").map(parseMatchupEntry);
}

export function parseDeckHistory(value: unknown): DeckHistoryEntry[] {
  return ensureArray(value, "deck history").map((item) => {
    const record = ensureRecord(item, "deck history entry");
    return {
      matchId: ensureString(record.match_id, "history match id"),
      opponentClass: optionalString(record.opponent_class),
      result: optionalString(record.result),
      playedAt: optionalString(record.played_at),
    };
  });
}
