import type { DeckCard } from "@/shared/model/dashboard";
import { Panel } from "@/shared/ui/panel";
import { cn } from "@/shared/lib/cn";

const RARITY_BACKGROUNDS: Record<string, string> = {
  LEGENDARY: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(36, 23, 14, 0.96) 58%, rgba(201, 133, 35, 0.92) 100%)",
  EPIC: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(32, 20, 49, 0.96) 58%, rgba(124, 73, 214, 0.9) 100%)",
  RARE: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(17, 31, 56, 0.96) 58%, rgba(33, 115, 216, 0.92) 100%)",
  COMMON: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(26, 33, 49, 0.96) 58%, rgba(101, 111, 131, 0.86) 100%)",
  FREE: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(24, 31, 46, 0.96) 58%, rgba(79, 92, 115, 0.84) 100%)",
  UNKNOWN: "linear-gradient(90deg, rgba(14, 18, 31, 0.98) 0%, rgba(24, 31, 46, 0.96) 58%, rgba(66, 77, 100, 0.84) 100%)",
};

export function DeckCardsGrid({ cards, loading }: { cards: DeckCard[]; loading: boolean }) {
  return (
    <Panel className="space-y-4">
      <div className="text-xl font-semibold text-app-text">Карты в колоде</div>
      {loading ? <div className="text-sm text-app-muted">Загрузка карт...</div> : null}
      {!loading && cards.length === 0 ? <div className="text-sm text-app-muted">Колода пока не содержит карт.</div> : null}
      <div className="grid gap-2 lg:grid-cols-2 xl:grid-cols-3">
        {cards.map((card) => (
          <div
            key={card.cardId}
            className={cn(
              "grid min-h-[var(--overlay-card-row-height)] grid-cols-[42px_1fr_40px] items-center overflow-hidden rounded-[var(--overlay-card-row-radius)] border border-slate-800/80 shadow-panel",
              card.isLegendary ? "ring-1 ring-amber-500/35" : "",
            )}
            style={{
              backgroundImage: buildDeckCardBackground(card),
              backgroundPosition: card.image ? "center 25%" : "center",
              backgroundSize: "cover",
            }}
          >
            <div className="grid h-full place-items-center bg-black/55 text-sm font-bold text-app-accent">{card.manaCost}</div>
            <div className="min-w-0 px-3">
              <div className="truncate text-sm font-semibold text-app-text drop-shadow-[0_1px_1px_rgba(0,0,0,0.7)]">{card.name}</div>
              <div className="text-[11px] uppercase tracking-[0.14em] text-app-muted">{card.rarity || "COMMON"}</div>
            </div>
            <div className="grid h-full place-items-center bg-black/55 text-sm font-black text-app-text">{card.copies}</div>
          </div>
        ))}
      </div>
    </Panel>
  );
}

function buildDeckCardBackground(card: DeckCard): string {
  const rarity = card.rarity.toUpperCase();
  const base = RARITY_BACKGROUNDS[rarity] ?? RARITY_BACKGROUNDS.UNKNOWN;

  if (!card.image) {
    return base;
  }

  const tint = card.isLegendary
    ? "linear-gradient(90deg, rgba(5, 8, 17, 0.86), rgba(24, 17, 12, 0.72) 58%, rgba(255, 183, 55, 0.32) 100%)"
    : "linear-gradient(90deg, rgba(5, 8, 17, 0.88), rgba(14, 19, 31, 0.76) 58%, rgba(117, 133, 161, 0.2) 100%)";

  return `${tint}, url(${card.image})`;
}
