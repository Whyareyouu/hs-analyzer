import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { AppProviders } from "@/app/providers/app-providers";
import { LiveOverlayControlPage } from "./page";

vi.mock("@/shared/config/app-config", () => ({
  getNativeBridge: () => ({
    getConfig: async () => ({ backendBaseUrl: "http://127.0.0.1:8080", liveBaseUrl: "http://127.0.0.1:8765" }),
    getOverlayMode: async () => "inspect",
    setOverlayMode: async (mode: "passive" | "inspect") => mode,
    getOverlayDiagnostics: async () => ({
      attached: true,
      degraded: false,
      gameDetected: true,
      reason: "ok",
      hearthstoneMode: "windowed",
      bounds: { x: 0, y: 0, width: 100, height: 100 },
    }),
    onOverlayDiagnostics: () => () => undefined,
  }),
}));

const snapshotPayload = {
  schema_version: "overlay_snapshot/v1",
  match: { match_id: "game:1", game_id: "1", status: "active", deck_id: "deck-1", deck_name: "Spell DH" },
  player_deck_remaining: [
    {
      card_id: "CARD_1",
      name: "Taste of Chaos",
      description: "Card text",
      image: "",
      mana_cost: 1,
      rarity: "COMMON",
      card_type: "SPELL",
      is_legendary: false,
      initial_copies: 2,
      remaining_copies: 1,
      drawn_copies: 1,
      played_copies: 0,
    },
  ],
  opponent_played_cards: [
    {
      card_id: "CARD_2",
      name: "Fireball",
      description: "Deal damage",
      image: "",
      mana_cost: 4,
      play_count: 1,
      first_played_at: "",
      last_played_at: "",
    },
  ],
  health: { monitor: "healthy", transport: "healthy" },
};

describe("LiveOverlayControlPage", () => {
  it("renders live snapshot from local transport", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      json: async () => snapshotPayload,
    }));
    vi.stubGlobal("fetch", fetchMock);
    class MockSocket {
      addEventListener() {}
      close() {}
    }
    vi.stubGlobal("WebSocket", MockSocket as unknown as typeof WebSocket);

    render(
      <AppProviders>
        <LiveOverlayControlPage />
      </AppProviders>,
    );

    await waitFor(() => expect(screen.getAllByText("Taste of Chaos")).toHaveLength(2));
    expect(screen.getByText("In deck (1)")).toBeInTheDocument();
    expect(screen.getByText("In hand (1)")).toBeInTheDocument();
    expect(screen.getByText("Fireball")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalled();
  });
});
