import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { DeckCardsGrid } from "./deck-cards-grid";

describe("DeckCardsGrid", () => {
  it("supports loading and empty states", () => {
    const { rerender } = render(<DeckCardsGrid cards={[]} loading />);
    expect(screen.getByText(/Загрузка карт/i)).toBeInTheDocument();

    rerender(<DeckCardsGrid cards={[]} loading={false} />);
    expect(screen.getByText(/не содержит карт/i)).toBeInTheDocument();
  });

  it("renders dashboard cards in the same strip style as overlay rows", () => {
    render(
      <DeckCardsGrid
        loading={false}
        cards={[
          {
            cardId: "CARD_1",
            name: "Taste of Chaos",
            manaCost: 1,
            copies: 2,
            rarity: "LEGENDARY",
            isLegendary: true,
            image: "https://example.com/card.png",
          },
        ]}
      />,
    );

    const row = screen.getByText("Taste of Chaos").closest("div[style]");
    expect(row).toHaveStyle({
      backgroundPosition: "center 25%",
      backgroundSize: "cover",
    });
    expect(screen.getByText("2")).toBeInTheDocument();
  });
});
