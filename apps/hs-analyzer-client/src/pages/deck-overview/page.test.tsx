import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { AppProviders } from "@/app/providers/app-providers";
import { DeckOverviewPage } from "./page";

vi.mock("@/shared/config/app-config", () => ({
  getNativeBridge: () => ({
    getConfig: async () => ({ backendBaseUrl: "http://127.0.0.1:8080", liveBaseUrl: "http://127.0.0.1:8765" }),
    getOverlayMode: async () => "passive",
    setOverlayMode: async (mode: "passive" | "inspect") => mode,
    getOverlayDiagnostics: async () => ({
      attached: false,
      degraded: true,
      gameDetected: false,
      reason: "n/a",
      hearthstoneMode: "unsupported",
      bounds: null,
    }),
    onOverlayDiagnostics: () => () => undefined,
  }),
}));

describe("DeckOverviewPage", () => {
  it("renders deck list and selected dashboard bundle", async () => {
    const fetchMock = vi.fn(async (input: string | URL) => {
      const url = String(input);
      if (url.endsWith("/api/v1/decks")) {
        return {
          ok: true,
          json: async () => [
            {
              deck_id: "deck-spell-dh",
              name: "Spell Demon Hunter",
              class: "DEMONHUNTER",
              wins: 18,
              losses: 12,
              winrate: 0.6,
              last_played_at: "2026-04-27T16:30:00Z",
              icon: "DH",
            },
          ],
        };
      }
      if (url.endsWith("/overview")) {
        return {
          ok: true,
          json: async () => ({
            deck_id: "deck-spell-dh",
            name: "Spell Demon Hunter",
            class: "DEMONHUNTER",
            wins: 18,
            losses: 12,
            total_matches: 30,
            winrate: 0.6,
            best_matchups: [{ opponent_class: "WARRIOR", wins: 6, losses: 2, winrate: 0.75, matches: 8 }],
            worst_matchups: [{ opponent_class: "MAGE", wins: 1, losses: 4, winrate: 0.2, matches: 5 }],
          }),
        };
      }
      if (url.endsWith("/cards")) {
        return {
          ok: true,
          json: async () => [
            {
              card_id: "CARD_1",
              name: "Taste of Chaos",
              mana_cost: 1,
              copies: 2,
              rarity: "COMMON",
              is_legendary: false,
              image: "",
            },
          ],
        };
      }
      return {
        ok: true,
        json: async () => [{ opponent_class: "WARRIOR", wins: 6, losses: 2, winrate: 0.75, matches: 8 }],
      };
    });
    vi.stubGlobal("fetch", fetchMock);

    render(
      <AppProviders>
        <DeckOverviewPage />
      </AppProviders>,
    );

    await waitFor(() => expect(screen.getByText("Карты в колоде")).toBeInTheDocument());
    expect(screen.getAllByText("Spell Demon Hunter")).toHaveLength(2);
    expect(screen.getByText("Taste of Chaos")).toBeInTheDocument();
  });
});
