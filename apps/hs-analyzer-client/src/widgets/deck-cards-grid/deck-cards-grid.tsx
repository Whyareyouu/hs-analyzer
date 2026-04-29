import type { DeckCard } from "@/shared/model/dashboard";
import { Panel } from "@/shared/ui/panel";
import { cn } from "@/shared/lib/cn";

export function DeckCardsGrid({ cards, loading }: { cards: DeckCard[]; loading: boolean }) {
  return (
    <Panel className="space-y-4">
      <div className="text-xl font-semibold text-app-text">Карты в колоде</div>
      {loading ? <div className="text-sm text-app-muted">Загрузка карт...</div> : null}
      {!loading && cards.length === 0 ? <div className="text-sm text-app-muted">Колода пока не содержит карт.</div> : null}
      <div className="grid gap-3 lg:grid-cols-2 xl:grid-cols-3">
        {cards.map((card) => (
          <div
            key={card.cardId}
            className={cn(
              "flex min-h-16 items-center gap-3 rounded-lg border border-app-border/50 bg-app-panelStrong/70 px-3 py-3",
              card.isLegendary ? "ring-1 ring-app-warning/50" : "",
            )}
          >
            <div className="grid h-10 w-10 place-items-center rounded-md bg-app-accent/15 text-sm font-bold text-app-accent">{card.manaCost}</div>
            <div className="min-w-0 flex-1">
              <div className="truncate text-sm font-semibold text-app-text">{card.name}</div>
              <div className="text-xs text-app-muted">{card.rarity || "COMMON"}</div>
            </div>
            <div className="text-sm font-semibold text-app-text">x{card.copies}</div>
          </div>
        ))}
      </div>
    </Panel>
  );
}
