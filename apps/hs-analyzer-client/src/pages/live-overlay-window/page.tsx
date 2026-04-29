import { Pill } from "@/shared/ui/pill";
import { useLiveOverlayData } from "@/features/reconnect-live-transport/model";
import { PlayerDeckPanel } from "@/widgets/overlay-player-deck/player-deck-panel";
import { OpponentCardsPanel } from "@/widgets/overlay-opponent-cards/opponent-cards-panel";

export function LiveOverlayWindowPage() {
  const { snapshot, status } = useLiveOverlayData();

  return (
    <div className="min-h-screen px-[var(--overlay-left-offset)] py-[var(--overlay-top-offset)]">
      <div className="mb-3 flex justify-center">
        {status !== "connected" ? <Pill tone="warning">overlay transport {status}</Pill> : null}
      </div>
      <div className="flex items-start justify-between gap-[var(--overlay-panel-gap)]">
        <PlayerDeckPanel cards={snapshot?.playerDeckRemaining ?? []} inspectMode={false} onInspect={() => undefined} />
        <OpponentCardsPanel cards={snapshot?.opponentPlayedCards ?? []} inspectMode={false} onInspect={() => undefined} />
      </div>
    </div>
  );
}
