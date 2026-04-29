import type { DashboardBundle, DeckCard, DeckListItem, DeckOverview, MatchupEntry } from "@/shared/model/dashboard";
import { parseDeckCards, parseDeckList, parseDeckOverview, parseMatchups } from "./decoders";

export class DashboardClient {
  private readonly cache = new Map<string, unknown>();

  constructor(private readonly baseUrl: string) {}

  async listDecks(): Promise<{ items: DeckListItem[]; stale: boolean }> {
    return this.fetchCached("/api/v1/decks", parseDeckList).then((result) => ({ items: result.data, stale: result.stale }));
  }

  async loadDeckBundle(deckId: string): Promise<DashboardBundle> {
    const [overview, cards, matchups] = await Promise.all([
      this.fetchCached<DeckOverview>(`/api/v1/decks/${deckId}/overview`, parseDeckOverview),
      this.fetchCached<DeckCard[]>(`/api/v1/decks/${deckId}/cards`, parseDeckCards),
      this.fetchCached<MatchupEntry[]>(`/api/v1/decks/${deckId}/matchups`, parseMatchups),
    ]);

    return {
      overview: overview.data,
      cards: cards.data,
      matchups: matchups.data,
      stale: overview.stale || cards.stale || matchups.stale,
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
