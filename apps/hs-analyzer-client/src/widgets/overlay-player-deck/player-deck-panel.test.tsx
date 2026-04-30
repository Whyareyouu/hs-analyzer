import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { PlayerDeckPanel } from "./player-deck-panel";

describe("PlayerDeckPanel", () => {
  it("splits the player deck into deck, hand, and played sections", () => {
    render(
      <PlayerDeckPanel
        otherCards={[
          {
            cardId: "OTHER_CARD",
            name: "Generated Spell",
            description: "",
            image: "",
            manaCost: 4,
            playCount: 2,
            firstPlayedAt: "",
            lastPlayedAt: "",
          },
        ]}
        inspectMode
        onInspect={vi.fn()}
        cards={[
          {
            cardId: "DECK_CARD",
            name: "Still In Deck",
            description: "",
            image: "",
            manaCost: 1,
            rarity: "RARE",
            cardType: "SPELL",
            isLegendary: false,
            initialCopies: 2,
            remainingCopies: 1,
            drawnCopies: 0,
            playedCopies: 0,
          },
          {
            cardId: "HAND_CARD",
            name: "Held Card",
            description: "",
            image: "",
            manaCost: 2,
            rarity: "COMMON",
            cardType: "MINION",
            isLegendary: false,
            initialCopies: 2,
            remainingCopies: 1,
            drawnCopies: 1,
            playedCopies: 0,
          },
          {
            cardId: "LEGENDARY_CARD",
            name: "Big Legend",
            description: "",
            image: "",
            manaCost: 7,
            rarity: "LEGENDARY",
            cardType: "MINION",
            isLegendary: true,
            initialCopies: 1,
            remainingCopies: 1,
            drawnCopies: 0,
            playedCopies: 0,
          },
        ]}
      />,
    );

    expect(screen.getByText("In deck (3)")).toBeInTheDocument();
    expect(screen.getByText("In hand (1)")).toBeInTheDocument();
    expect(screen.getByText("Other (2)")).toBeInTheDocument();
    expect(screen.getByText("Still In Deck")).toBeInTheDocument();
    expect(screen.getAllByText("Held Card")).toHaveLength(2);
    expect(screen.getByText("Generated Spell")).toBeInTheDocument();
    expect(screen.getAllByLabelText("1 copies")).toHaveLength(4);
    expect(screen.getByLabelText("2 copies")).toHaveTextContent("2");
    expect(screen.getAllByLabelText("1 copies")[0]).toHaveTextContent("1");
  });

  it("shows explicit waiting state when deck cards are not available yet", () => {
    render(<PlayerDeckPanel inspectMode={false} onInspect={vi.fn()} cards={[]} otherCards={[]} />);
    expect(screen.getByText(/Ждём выбранную колоду/i)).toBeInTheDocument();
  });

  it("anchors card art to the top of a taller row when image is available", () => {
    render(
      <PlayerDeckPanel
        otherCards={[]}
        inspectMode={false}
        onInspect={vi.fn()}
        cards={[
          {
            cardId: "ART_CARD",
            name: "Art Card",
            description: "",
            image: "https://example.com/card.png",
            manaCost: 5,
            rarity: "RARE",
            cardType: "MINION",
            isLegendary: false,
            initialCopies: 1,
            remainingCopies: 1,
            drawnCopies: 0,
            playedCopies: 0,
          },
        ]}
      />,
    );

    expect(screen.getByRole("button", { name: /Art Card/i })).toHaveStyle({
      backgroundPosition: "center 25%",
      backgroundSize: "cover",
      minHeight: "var(--overlay-card-row-height)",
    });
  });

  it("keeps long player lists scrollable inside the panel", () => {
    const cards = Array.from({ length: 16 }, (_, index) => ({
      cardId: `CARD_${index}`,
      name: `Deck Card ${index}`,
      description: "",
      image: "",
      manaCost: index,
      rarity: "COMMON",
      cardType: "MINION",
      isLegendary: false,
      initialCopies: 1,
      remainingCopies: 1,
      drawnCopies: 0,
      playedCopies: 0,
    }));

    const { container } = render(<PlayerDeckPanel inspectMode={false} onInspect={vi.fn()} cards={cards} otherCards={[]} />);

    const panel = container.firstElementChild;
    expect(panel).toHaveStyle({ maxHeight: "calc(100vh - (var(--overlay-top-offset) * 2))" });
    expect(panel?.firstElementChild).toHaveClass("overflow-y-auto");
  });
});
