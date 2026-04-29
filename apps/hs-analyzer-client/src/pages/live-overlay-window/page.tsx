import { useLiveOverlayData } from "@/features/reconnect-live-transport/model";
import { PlayerDeckPanel } from "@/widgets/overlay-player-deck/player-deck-panel";
import { OpponentCardsPanel } from "@/widgets/overlay-opponent-cards/opponent-cards-panel";

export function LiveOverlayWindowPage() {
  const { snapshot } = useLiveOverlayData();

  return (
    <div className="relative h-screen overflow-hidden">
      <div className="absolute left-[var(--overlay-left-offset)] top-[var(--overlay-top-offset)]">
        <OpponentCardsPanel cards={snapshot?.opponentPlayedCards ?? []} inspectMode={false} onInspect={() => undefined} />
      </div>
      <div className="absolute right-[var(--overlay-right-offset)] top-[var(--overlay-top-offset)]">
        <PlayerDeckPanel cards={snapshot?.playerDeckRemaining ?? []} inspectMode={false} onInspect={() => undefined} />
      </div>
    </div>
  );
}
