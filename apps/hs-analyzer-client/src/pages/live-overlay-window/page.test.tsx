import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { AppProviders } from "@/app/providers/app-providers";
import { LiveOverlayWindowPage } from "./page";

const bridgeState = vi.hoisted(() => ({
  setOverlayInspectInteractivity: vi.fn(async (_active: boolean) => undefined),
}));

vi.mock("@/shared/config/app-config", () => ({
  getNativeBridge: () => ({
    getConfig: async () => ({ backendBaseUrl: "http://127.0.0.1:8080", liveBaseUrl: "http://127.0.0.1:8765" }),
    getOverlayMode: async () => "inspect",
    setOverlayMode: async (mode: "passive" | "inspect") => mode,
    setOverlayInspectInteractivity: bridgeState.setOverlayInspectInteractivity,
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
      playerOtherPlayedCards: [],
      opponentPlayedCards: [
        {
          cardId: "OPP_CARD",
          name: "Opponent Card",
          description: "",
          image: "https://example.com/opponent-card.png",
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
    bridgeState.setOverlayInspectInteractivity.mockClear();
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
    fireEvent.pointerMove(window, { clientX: 8, clientY: 8 });
    await waitFor(() => expect(bridgeState.setOverlayInspectInteractivity).toHaveBeenCalledWith(false));

    const opponentCard = screen.getByRole("button", { name: /Opponent Card/i });
    vi.spyOn(opponentCard, "getBoundingClientRect").mockReturnValue({
      x: 24,
      y: 180,
      width: 300,
      height: 26,
      top: 180,
      right: 324,
      bottom: 206,
      left: 24,
      toJSON: () => undefined,
    });

    fireEvent.mouseEnter(opponentCard);
    expect(screen.getByRole("img", { name: "Opponent Card" })).toBeInTheDocument();
    expect(screen.queryByText("Описание карты недоступно.")).not.toBeInTheDocument();
    expect(screen.getByTestId("overlay-card-tooltip")).toHaveStyle({ top: "180px" });

    fireEvent.mouseLeave(opponentCard);

    const myCard = screen.getByRole("button", { name: /My Card/i });
    vi.spyOn(myCard, "getBoundingClientRect").mockReturnValue({
      x: 900,
      y: 220,
      width: 300,
      height: 26,
      top: 220,
      right: 1200,
      bottom: 246,
      left: 900,
      toJSON: () => undefined,
    });

    fireEvent.mouseEnter(myCard);
    expect(screen.queryByRole("img", { name: "My Card" })).not.toBeInTheDocument();
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
