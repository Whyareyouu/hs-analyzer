export interface DeckListItem {
  deckId: string;
  name: string;
  className: string;
  wins: number;
  losses: number;
  winrate: number;
  lastPlayedAt: string;
  icon: string;
}

export interface MatchupEntry {
  opponentClass: string;
  opponentArchetype: string;
  wins: number;
  losses: number;
  winrate: number;
  matches: number;
}

export interface DeckOverview {
  deckId: string;
  name: string;
  className: string;
  wins: number;
  losses: number;
  totalMatches: number;
  winrate: number;
  bestMatchups: MatchupEntry[];
  worstMatchups: MatchupEntry[];
}

export interface DeckCard {
  cardId: string;
  name: string;
  manaCost: number;
  copies: number;
  rarity: string;
  isLegendary: boolean;
  image: string;
}

export interface DeckHistoryEntry {
  matchId: string;
  opponentClass: string;
  result: string;
  playedAt: string;
}

export interface DashboardBundle {
  overview: DeckOverview;
  cards: DeckCard[];
  matchups: MatchupEntry[];
  history: DeckHistoryEntry[];
  stale: boolean;
}
