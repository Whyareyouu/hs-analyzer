import type { OverlayDeckCard } from "@/shared/model/overlay";
import { cn } from "@/shared/lib/cn";
import { Panel } from "@/shared/ui/panel";

export function PlayerDeckPanel({
  cards,
  inspectMode,
  onInspect,
}: {
  cards: OverlayDeckCard[];
  inspectMode: boolean;
  onInspect(card: OverlayDeckCard | null): void;
}) {
  return (
    <Panel className="w-[var(--overlay-panel-width)] p-3">
      <div className="mb-3 text-xs uppercase tracking-[0.2em] text-app-muted">Моя колода</div>
      <div className="space-y-2">
        {cards.length === 0 ? <div className="text-xs text-app-muted">Ждём выбранную колоду из Decks.log.</div> : null}
        {cards.map((card) => (
          <button
            key={card.cardId}
            type="button"
            onMouseEnter={() => inspectMode && onInspect(card)}
            onMouseLeave={() => onInspect(null)}
            onFocus={() => inspectMode && onInspect(card)}
            onBlur={() => onInspect(null)}
            className={cn(
              "grid w-[var(--overlay-card-row-width)] grid-cols-[36px_1fr_var(--overlay-card-count-width)] items-center overflow-hidden rounded-[var(--overlay-card-row-radius)] border border-app-border/60 text-left shadow-panel",
              card.remainingCopies === 0 ? "grayscale" : "opacity-100",
            )}
            style={{
              minHeight: "var(--overlay-card-row-height)",
              backgroundImage:
                card.remainingCopies === 0
                  ? "linear-gradient(90deg, rgba(8, 8, 10, 0.96), rgba(30, 32, 36, 0.92))"
                  : card.image
                    ? `linear-gradient(90deg, rgba(5, 10, 19, 0.9), rgba(5, 10, 19, 0.35)), url(${card.image})`
                    : "linear-gradient(90deg, rgba(12, 18, 31, 0.95), rgba(18, 28, 49, 0.7))",
              backgroundSize: "cover",
              backgroundPosition: "center",
              opacity: card.remainingCopies === 0 ? 0.78 : 1,
            }}
          >
            <span className="grid h-full place-items-center bg-black/35 text-xs font-bold text-app-accent">{card.manaCost}</span>
            <span className="truncate px-3 text-xs font-medium text-app-text">{card.name}</span>
            <span className="grid h-full place-items-center bg-black/30 text-[10px] font-semibold text-app-text">
              {card.remainingCopies}/{card.initialCopies}
            </span>
          </button>
        ))}
      </div>
    </Panel>
  );
}
