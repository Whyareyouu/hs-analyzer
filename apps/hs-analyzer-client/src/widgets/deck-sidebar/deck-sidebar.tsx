import type { DeckListItem } from "@/shared/model/dashboard";
import { Panel } from "@/shared/ui/panel";
import { Pill } from "@/shared/ui/pill";
import { cn } from "@/shared/lib/cn";
import { formatRecord, formatWinrate } from "@/shared/lib/format";

export function DeckSidebar({
  decks,
  selectedDeckId,
  onSelect,
}: {
  decks: DeckListItem[];
  selectedDeckId: string;
  onSelect(deckId: string): void;
}) {
  return (
    <Panel className="flex h-full flex-col gap-4">
      <div className="flex items-center justify-between">
        <div>
          <div className="text-xs uppercase tracking-[0.24em] text-app-muted">Dashboard</div>
          <div className="text-2xl font-bold text-app-text">Мои колоды</div>
        </div>
        <Pill>{decks.length}</Pill>
      </div>
      <div className="flex flex-col gap-3">
        {decks.map((deck) => {
          const selected = deck.deckId === selectedDeckId;
          return (
            <button
              key={deck.deckId}
              type="button"
              onClick={() => onSelect(deck.deckId)}
              className={cn(
                "flex min-h-[var(--dashboard-deck-row-min-height)] items-center gap-3 rounded-lg border px-4 py-3 text-left transition-all duration-base focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-accent/60",
                selected
                  ? "border-app-accent/70 bg-app-accent/15 shadow-glow"
                  : "border-app-border/50 bg-app-panel/40 hover:border-app-border hover:bg-app-panelStrong/70",
              )}
            >
              <div className="grid h-10 w-10 place-items-center rounded-full bg-app-panelStrong text-sm font-bold text-app-accent">{deck.icon || deck.className.slice(0, 2)}</div>
              <div className="min-w-0 flex-1">
                <div className="truncate text-sm font-semibold text-app-text">{deck.name}</div>
                <div className="text-xs text-app-muted">{formatRecord(deck.wins, deck.losses)}</div>
              </div>
              <div className="text-sm font-semibold text-app-text">{formatWinrate(deck.winrate)}</div>
            </button>
          );
        })}
      </div>
    </Panel>
  );
}
