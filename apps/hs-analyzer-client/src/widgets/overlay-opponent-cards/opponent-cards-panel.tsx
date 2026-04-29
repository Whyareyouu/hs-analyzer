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
        {cards.map((card) => (
          <button
            key={card.cardId}
            type="button"
            onMouseEnter={() => inspectMode && onInspect(card)}
            onMouseLeave={() => onInspect(null)}
            onFocus={() => inspectMode && onInspect(card)}
            onBlur={() => onInspect(null)}
            className="grid w-[var(--overlay-card-row-width)] grid-cols-[36px_1fr_var(--overlay-card-count-width)] items-center overflow-hidden rounded-[var(--overlay-card-row-radius)] border border-app-border/60 bg-[linear-gradient(90deg,rgba(30,10,12,0.92),rgba(20,25,40,0.8))] text-left shadow-panel"
            style={{ minHeight: "var(--overlay-card-row-height)" }}
          >
            <span className="grid h-full place-items-center bg-black/35 text-xs font-bold text-app-accent">{card.manaCost}</span>
            <span className="truncate px-3 text-xs font-medium text-app-text">{card.name}</span>
            <span className="grid h-full place-items-center bg-black/30 text-xs font-semibold text-app-text">x{card.playCount}</span>
          </button>
        ))}
      </div>
    </Panel>
  );
}
