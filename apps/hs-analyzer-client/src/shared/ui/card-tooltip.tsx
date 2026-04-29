import type { OverlayDeckCard, OpponentPlayedCard } from "@/shared/model/overlay";
import { Panel } from "./panel";

type InspectableCard = OverlayDeckCard | OpponentPlayedCard;

export function CardTooltip({ card }: { card: InspectableCard | null }) {
  if (!card) {
    return null;
  }

  return (
    <Panel className="w-[var(--overlay-tooltip-width)] p-4">
      <div className="flex gap-4">
        <div
          className="h-28 w-20 shrink-0 overflow-hidden rounded-md border border-app-border/70 bg-app-panelStrong bg-cover bg-center"
          style={card.image ? { backgroundImage: `url(${card.image})` } : undefined}
          aria-hidden="true"
        >
          {!card.image ? <div className="grid h-full place-items-center text-xs text-app-muted">No art</div> : null}
        </div>
        <div className="space-y-2">
          <div className="text-base font-semibold text-app-text">{card.name}</div>
          <div className="text-sm leading-6 text-app-muted">{card.description || "Описание карты недоступно."}</div>
        </div>
      </div>
    </Panel>
  );
}
