import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { AppProviders } from "@/app/providers/app-providers";
import { LiveOverlayWindowPage } from "./page";

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

const overlayFixture = {
  status: "active" as "idle" | "active" | "finalized",
};

vi.mock("@/features/reconnect-live-transport/model", () => ({
  useLiveOverlayData: () => ({
    status: "disconnected",
    snapshot: {
      schemaVersion: "overlay_snapshot/v1",
      match: { status: overlayFixture.status, deckId: "deck-1", deckName: "My Deck" },
      playerDeckRemaining: [
        {
          cardId: "SELF_CARD",
          name: "My Card",
          description: "",
          image: "",
          manaCost: 1,
          rarity: "COMMON",
          cardType: "MINION",
          isLegendary: false,
          initialCopies: 2,
          remainingCopies: 2,
          drawnCopies: 0,
          playedCopies: 0,
        },
      ],
      opponentPlayedCards: [
        {
          cardId: "OPP_CARD",
          name: "Opponent Card",
          description: "",
          image: "",
          manaCost: 4,
          playCount: 1,
          firstPlayedAt: "",
          lastPlayedAt: "",
        },
      ],
      health: { monitor: "healthy", transport: "healthy" },
    },
  }),
}));

describe("LiveOverlayWindowPage", () => {
  it("renders opponent cards and player deck with side tooltips in inspect mode", async () => {
    overlayFixture.status = "active";

    const { container } = render(
      <AppProviders>
        <LiveOverlayWindowPage />
      </AppProviders>,
    );

    expect(screen.queryByText(/overlay transport/i)).not.toBeInTheDocument();
    expect(screen.getByText("Оппонент сыграл")).toBeInTheDocument();
    expect(screen.getByText("In deck (2)")).toBeInTheDocument();

    const panels = container.querySelectorAll(".absolute");
    expect(panels).toHaveLength(2);
    expect(panels[0]).toHaveTextContent("Opponent Card");
    expect(panels[0]).not.toHaveTextContent("My Card");
    expect(panels[1]).toHaveTextContent("My Card");
    expect(panels[1]).not.toHaveTextContent("Opponent Card");

    await waitFor(() => expect(screen.getByRole("button", { name: /Opponent Card/i })).toBeInTheDocument());

    fireEvent.mouseEnter(screen.getByRole("button", { name: /Opponent Card/i }));
    expect(screen.getByText("Описание карты недоступно.")).toBeInTheDocument();

    fireEvent.mouseLeave(screen.getByRole("button", { name: /Opponent Card/i }));
    fireEvent.mouseEnter(screen.getByRole("button", { name: /My Card/i }));
    expect(screen.getAllByText("My Card")).toHaveLength(2);
  });

  it("keeps the game overlay hidden until a match becomes active", () => {
    cleanup();
    overlayFixture.status = "idle";

    render(
      <AppProviders>
        <LiveOverlayWindowPage />
      </AppProviders>,
    );

    expect(screen.queryByText("Оппонент сыграл")).not.toBeInTheDocument();
    expect(screen.queryByText(/In deck/i)).not.toBeInTheDocument();
  });
});
