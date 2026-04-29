import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { LiveOverlayWindowPage } from "./page";

vi.mock("@/features/reconnect-live-transport/model", () => ({
  useLiveOverlayData: () => ({
    status: "disconnected",
    snapshot: {
      schemaVersion: "overlay_snapshot/v1",
      match: { status: "active", deckId: "deck-1", deckName: "My Deck" },
      playerDeckRemaining: [
        {
          cardId: "SELF_CARD",
          name: "My Card",
          description: "",
          image: "",
          manaCost: 1,
          initialCopies: 2,
          remainingCopies: 2,
          drawnCopies: 0,
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
  it("renders opponent played cards on the left and player deck on the right without a status strip", () => {
    const { container } = render(<LiveOverlayWindowPage />);

    expect(screen.queryByText(/overlay transport/i)).not.toBeInTheDocument();
    expect(screen.getByText("Оппонент сыграл")).toBeInTheDocument();
    expect(screen.getByText("Моя колода")).toBeInTheDocument();

    const panels = container.querySelectorAll(".absolute");
    expect(panels).toHaveLength(2);
    expect(panels[0]).toHaveTextContent("Opponent Card");
    expect(panels[0]).not.toHaveTextContent("My Card");
    expect(panels[1]).toHaveTextContent("My Card");
    expect(panels[1]).not.toHaveTextContent("Opponent Card");
  });
});
