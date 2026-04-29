import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { DeckSummary } from "./deck-summary";

describe("DeckSummary", () => {
  it("renders three-panel summary row with hero winrate ring", () => {
    const { container } = render(
      <DeckSummary
        stale={false}
        overview={{
          deckId: "deck-1",
          name: "Spell DH",
          className: "DEMONHUNTER",
          wins: 18,
          losses: 12,
          totalMatches: 30,
          winrate: 0.6,
          bestMatchups: [{ opponentClass: "WARRIOR", opponentArchetype: "", wins: 6, losses: 2, winrate: 0.75, matches: 8 }],
          worstMatchups: [],
        }}
      />,
    );

    expect(screen.getByText("Общая статистика")).toBeInTheDocument();
    expect(screen.getAllByText("60%")).toHaveLength(2);
    expect(container.firstChild).toMatchSnapshot();
  });
});
