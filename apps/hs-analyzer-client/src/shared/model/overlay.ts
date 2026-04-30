export interface OverlaySnapshot {
  schemaVersion: string;
  match: {
    matchId: string;
    gameId: string;
    status: "idle" | "active" | "finalized";
    playerClass: string;
    opponentClass: string;
    deckId: string;
    deckName: string;
  };
  playerDeckRemaining: OverlayDeckCard[];
  playerOtherPlayedCards: OpponentPlayedCard[];
  opponentPlayedCards: OpponentPlayedCard[];
  health: {
    monitor: "healthy" | "degraded";
    transport: "healthy" | "degraded";
  };
}

export interface OverlayDeckCard {
  cardId: string;
  name: string;
  description: string;
  image: string;
  manaCost: number;
  rarity: string;
  cardType: string;
  isLegendary: boolean;
  initialCopies: number;
  remainingCopies: number;
  drawnCopies: number;
  playedCopies: number;
}

export interface OpponentPlayedCard {
  cardId: string;
  name: string;
  description: string;
  image: string;
  manaCost: number;
  playCount: number;
  firstPlayedAt: string;
  lastPlayedAt: string;
}

export interface OverlayEnvelope {
  type: "overlay_snapshot" | "overlay_delta" | "overlay_reset" | "transport_health";
  schemaVersion: string;
  snapshot?: OverlaySnapshot;
  delta?: OverlaySnapshot;
  health?: OverlaySnapshot["health"];
  matchId?: string;
}
