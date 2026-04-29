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
});
