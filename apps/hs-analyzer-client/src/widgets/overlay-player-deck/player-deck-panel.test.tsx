import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { PlayerDeckPanel } from "./player-deck-panel";

describe("PlayerDeckPanel", () => {
  it("renders overlay rows with tokenized layout and zero-count state", () => {
    const { container } = render(
      <PlayerDeckPanel
        inspectMode
        onInspect={vi.fn()}
        cards={[
          {
            cardId: "CARD_1",
            name: "Taste of Chaos",
            description: "",
            image: "",
            manaCost: 1,
            initialCopies: 2,
            remainingCopies: 0,
            drawnCopies: 2,
          },
        ]}
      />,
    );

    expect(screen.getByText("Taste of Chaos")).toBeInTheDocument();
    expect(container.firstChild).toMatchSnapshot();
  });

  it("shows explicit waiting state when deck cards are not available yet", () => {
    render(<PlayerDeckPanel inspectMode={false} onInspect={vi.fn()} cards={[]} />);
    expect(screen.getByText(/Ждём выбранную колоду/i)).toBeInTheDocument();
  });
});
