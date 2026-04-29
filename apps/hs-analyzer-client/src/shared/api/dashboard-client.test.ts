import { describe, expect, it, vi } from "vitest";
import { DashboardClient } from "./dashboard-client";

describe("DashboardClient", () => {
  it("falls back to stale cache when backend becomes unavailable", async () => {
    const client = new DashboardClient("http://127.0.0.1:8080");
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [
          {
            deck_id: "deck-1",
            name: "Spell DH",
            class: "DEMONHUNTER",
            wins: 18,
            losses: 12,
            winrate: 0.6,
            last_played_at: "2026-04-27T16:30:00Z",
            icon: "DH",
          },
        ],
      })
      .mockRejectedValueOnce(new Error("backend offline"));
    vi.stubGlobal("fetch", fetchMock);

    const first = await client.listDecks();
    const second = await client.listDecks();

    expect(first.stale).toBe(false);
    expect(second.stale).toBe(true);
    expect(second.items[0]?.deckId).toBe("deck-1");
  });
});
