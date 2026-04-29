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
});
