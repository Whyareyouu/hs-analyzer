import { describe, expect, it } from "vitest";
import { parseOverlaySnapshot } from "@/shared/api/decoders";

describe("overlay payload performance", () => {
  it("parses a large snapshot within acceptable UI latency budget", () => {
    const raw = {
      schema_version: "overlay_snapshot/v1",
      match: { match_id: "game:1", game_id: "1", status: "active", deck_id: "deck-1", deck_name: "Deck" },
      player_deck_remaining: Array.from({ length: 30 }, (_, index) => ({
        card_id: `CARD_${index}`,
        name: `Card ${index}`,
        description: "Card text",
        image: "",
        mana_cost: index % 8,
        initial_copies: 2,
        remaining_copies: 1,
        drawn_copies: 1,
      })),
      opponent_played_cards: Array.from({ length: 20 }, (_, index) => ({
        card_id: `OPP_${index}`,
        name: `Opp ${index}`,
        description: "Opp text",
        image: "",
        mana_cost: index % 8,
        play_count: 2,
        first_played_at: "",
        last_played_at: "",
      })),
      health: { monitor: "healthy", transport: "healthy" },
    };

    const started = performance.now();
    for (let index = 0; index < 500; index += 1) {
      parseOverlaySnapshot(raw);
    }
    const elapsed = performance.now() - started;

    expect(elapsed).toBeLessThan(120);
  });
});
