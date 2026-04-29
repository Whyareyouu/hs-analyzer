import { describe, expect, it } from "vitest";
import { parseDeckList, parseOverlaySnapshot } from "./decoders";

describe("decoders", () => {
  it("accepts valid overlay snapshot payload", () => {
    const snapshot = parseOverlaySnapshot({
      schema_version: "overlay_snapshot/v1",
      match: { match_id: "game:1", game_id: "1", status: "active", deck_id: "deck-1", deck_name: "Deck" },
      player_deck_remaining: [
        {
          card_id: "CARD_1",
          name: "Card",
          description: "",
          image: "",
          mana_cost: 2,
          initial_copies: 2,
          remaining_copies: 1,
          drawn_copies: 1,
        },
      ],
      opponent_played_cards: [],
      health: { monitor: "healthy", transport: "healthy" },
    });

    expect(snapshot.match.deckId).toBe("deck-1");
    expect(snapshot.playerDeckRemaining[0].remainingCopies).toBe(1);
  });

  it("rejects invalid dashboard payload", () => {
    expect(() => parseDeckList([{ deck_id: 1 }])).toThrowError(/deck id/i);
  });
});
