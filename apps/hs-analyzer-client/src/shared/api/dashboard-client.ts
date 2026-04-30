import type { DashboardBundle, DeckCard, DeckHistoryEntry, DeckListItem, DeckOverview, MatchupEntry } from "@/shared/model/dashboard";
import { parseDeckCards, parseDeckHistory, parseDeckList, parseDeckOverview, parseMatchups } from "./decoders";

export class DashboardClient {
  private readonly cache = new Map<string, unknown>();

  constructor(private readonly baseUrl: string) {}

  async listDecks(): Promise<{ items: DeckListItem[]; stale: boolean }> {
    return this.fetchCached("/api/v1/decks", parseDeckList).then((result) => ({ items: result.data, stale: result.stale }));
  }

  async loadDeckBundle(deckId: string): Promise<DashboardBundle> {
    const encodedDeckId = encodeURIComponent(deckId);
    const [overview, cards, matchups, history] = await Promise.all([
      this.fetchCached<DeckOverview>(`/api/v1/decks/${encodedDeckId}/overview`, parseDeckOverview),
      this.fetchCached<DeckCard[]>(`/api/v1/decks/${encodedDeckId}/cards`, parseDeckCards),
      this.fetchCached<MatchupEntry[]>(`/api/v1/decks/${encodedDeckId}/matchups`, parseMatchups),
      this.fetchCached<DeckHistoryEntry[]>(`/api/v1/decks/${encodedDeckId}/history`, parseDeckHistory),
    ]);

    return {
      overview: overview.data,
      cards: cards.data,
      matchups: matchups.data,
      history: history.data,
      stale: overview.stale || cards.stale || matchups.stale || history.stale,
    };
  }

  private async fetchCached<T>(path: string, decoder: (value: unknown) => T): Promise<{ data: T; stale: boolean }> {
    try {
      const response = await fetch(`${this.baseUrl}${path}`);
      if (!response.ok) {
        throw new Error(`request failed with status ${response.status}`);
      }
      const data = decoder(await response.json());
      this.cache.set(path, data);
      return { data, stale: false };
    } catch (error) {
      if (this.cache.has(path)) {
        return { data: this.cache.get(path) as T, stale: true };
      }
      throw error;
    }
  }
}
