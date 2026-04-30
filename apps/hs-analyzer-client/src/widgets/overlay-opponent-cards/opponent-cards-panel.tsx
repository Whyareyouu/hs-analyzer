import type { OpponentPlayedCard } from "@/shared/model/overlay";
import { Panel } from "@/shared/ui/panel";

export function OpponentCardsPanel({
  cards,
  inspectMode,
  onInspect,
}: {
  cards: OpponentPlayedCard[];
  inspectMode: boolean;
  onInspect(card: OpponentPlayedCard | null): void;
}) {
  return (
    <Panel className="w-[var(--overlay-panel-width)] p-3">
      <div className="mb-3 text-xs uppercase tracking-[0.2em] text-app-muted">Оппонент сыграл</div>
      <div className="space-y-2">
        {cards.length === 0 ? <div className="text-xs text-app-muted">Оппонент пока ничего не разыграл.</div> : null}
        {cards.map((card) => (
          <button
            key={card.cardId}
            type="button"
            onMouseEnter={() => inspectMode && onInspect(card)}
            onMouseLeave={() => onInspect(null)}
            onFocus={() => inspectMode && onInspect(card)}
            onBlur={() => onInspect(null)}
            className="grid w-[var(--overlay-card-row-width)] grid-cols-[42px_1fr_var(--overlay-card-count-width)] items-center overflow-hidden rounded-[var(--overlay-card-row-radius)] border border-slate-900/90 text-left shadow-panel"
            style={{
              minHeight: "var(--overlay-card-row-height)",
              backgroundImage: buildOpponentRowBackground(card),
              backgroundPosition: card.image ? "center 25%" : "center",
              backgroundSize: "cover",
            }}
          >
            <span className="grid h-full place-items-center bg-black/55 text-xs font-bold text-app-accent">{card.manaCost}</span>
            <span className="truncate px-3 text-xs font-medium text-app-text drop-shadow-[0_1px_1px_rgba(0,0,0,0.7)]">{card.name}</span>
            <span className="grid h-full place-items-center bg-black/55 text-xs font-semibold text-app-text">x{card.playCount}</span>
          </button>
        ))}
      </div>
    </Panel>
  );
}

function buildOpponentRowBackground(card: OpponentPlayedCard): string {
  const base = "linear-gradient(90deg, rgba(6, 9, 18, 0.98) 0%, rgba(22, 14, 18, 0.96) 54%, rgba(91, 29, 39, 0.88) 100%)";
  if (!card.image) {
    return base;
  }
  return `linear-gradient(90deg, rgba(5, 8, 17, 0.88) 0%, rgba(18, 10, 14, 0.72) 54%, rgba(127, 34, 49, 0.22) 100%), url(${card.image})`;
}
