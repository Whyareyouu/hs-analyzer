import { useDeckSelection } from "@/features/select-deck/model";
import { DeckCardsGrid } from "@/widgets/deck-cards-grid/deck-cards-grid";
import { DeckSidebar } from "@/widgets/deck-sidebar/deck-sidebar";
import { DeckSummary } from "@/widgets/deck-summary/deck-summary";
import { MatchupBreakdown } from "@/widgets/matchup-breakdown/matchup-breakdown";

export function DeckOverviewPage() {
  const { decks, selectedDeckId, selectDeck, bundle, status, stale, errorMessage } = useDeckSelection();

  return (
    <div className="grid min-h-screen gap-[var(--dashboard-panel-gap)] p-6 xl:grid-cols-[var(--dashboard-sidebar-width)_1fr]">
      <DeckSidebar decks={decks} selectedDeckId={selectedDeckId} onSelect={selectDeck} />
      <div className="space-y-[var(--dashboard-panel-gap)]">
        {status === "error" ? <div className="rounded-xl border border-app-danger/50 bg-app-danger/10 p-4 text-sm text-app-danger">Не удалось загрузить dashboard: {errorMessage}</div> : null}
        {bundle ? (
          <>
            <DeckSummary overview={bundle.overview} stale={stale || bundle.stale} />
            <MatchupBreakdown matchups={bundle.matchups} />
            <DeckCardsGrid cards={bundle.cards} loading={status === "loading"} />
          </>
        ) : (
          <div className="rounded-xl border border-app-border/70 bg-app-panel/50 p-6 text-sm text-app-muted">Загрузка dashboard...</div>
        )}
      </div>
    </div>
  );
}
