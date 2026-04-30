import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { OpponentCardsPanel } from "./opponent-cards-panel";

describe("OpponentCardsPanel", () => {
  it("renders played cards with counts", () => {
    render(
      <OpponentCardsPanel
        inspectMode
        onInspect={vi.fn()}
        cards={[
          {
            cardId: "CARD_2",
            name: "Fireball",
            description: "",
            image: "",
            manaCost: 4,
            playCount: 2,
            firstPlayedAt: "",
            lastPlayedAt: "",
          },
        ]}
      />,
    );

    expect(screen.getByText("Fireball")).toBeInTheDocument();
    expect(screen.getByText("x2")).toBeInTheDocument();
  });

  it("shows explicit waiting state before the opponent plays cards", () => {
    render(<OpponentCardsPanel inspectMode={false} onInspect={vi.fn()} cards={[]} />);
    expect(screen.getByText(/Оппонент пока ничего не разыграл/i)).toBeInTheDocument();
  });

  it("uses the top of the card art as the row background when image is available", () => {
    render(
      <OpponentCardsPanel
        inspectMode={false}
        onInspect={vi.fn()}
        cards={[
          {
            cardId: "CARD_3",
            name: "Frostbolt",
            description: "",
            image: "https://example.com/frostbolt.png",
            manaCost: 2,
            playCount: 1,
            firstPlayedAt: "",
            lastPlayedAt: "",
          },
        ]}
      />,
    );

    expect(screen.getByRole("button", { name: /Frostbolt/i })).toHaveStyle({
      backgroundPosition: "center 25%",
      backgroundSize: "cover",
      minHeight: "var(--overlay-card-row-height)",
    });
  });

  it("keeps long opponent lists scrollable inside the panel", () => {
    const cards = Array.from({ length: 18 }, (_, index) => ({
      cardId: `CARD_${index}`,
      name: `Card ${index}`,
      description: "",
      image: "",
      manaCost: index,
      playCount: 1,
      firstPlayedAt: "",
      lastPlayedAt: "",
    }));

    const { container } = render(<OpponentCardsPanel inspectMode={false} onInspect={vi.fn()} cards={cards} />);

    const panel = container.firstElementChild;
    expect(panel).toHaveStyle({ maxHeight: "calc(100vh - (var(--overlay-top-offset) * 2))" });
    expect(panel?.lastElementChild).toHaveClass("overflow-y-auto");
  });
});
