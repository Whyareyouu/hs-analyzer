import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { DeckSidebar } from "./deck-sidebar";

describe("DeckSidebar", () => {
  it("shows pronounced selected state and updates selection", async () => {
    const onSelect = vi.fn();
    render(
      <DeckSidebar
        selectedDeckId="deck-1"
        onSelect={onSelect}
        decks={[
          {
            deckId: "deck-1",
            name: "Spell DH",
            className: "DEMONHUNTER",
            wins: 18,
            losses: 12,
            winrate: 0.6,
            lastPlayedAt: "",
            icon: "DH",
          },
          {
            deckId: "deck-2",
            name: "Rainbow DK",
            className: "DEATHKNIGHT",
            wins: 11,
            losses: 9,
            winrate: 0.55,
            lastPlayedAt: "",
            icon: "DK",
          },
        ]}
      />,
    );

    await userEvent.click(screen.getByRole("button", { name: /Rainbow DK/i }));
    expect(onSelect).toHaveBeenCalledWith("deck-2");
  });

  it("renders empty-state copy when decks are unavailable", () => {
    render(<DeckSidebar selectedDeckId="" onSelect={() => undefined} decks={[]} />);
    expect(screen.getByText(/Не удалось найти колоды/i)).toBeInTheDocument();
  });
});
