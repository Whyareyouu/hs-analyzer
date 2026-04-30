import { useEffect, useState } from "react";
import type { OverlayMode } from "@electron/shared/contracts";
import { useServices } from "@/shared/api/services-context";
import { CardTooltip } from "@/shared/ui/card-tooltip";
import { useCardInspectState } from "@/features/inspect-card/model";
import { useLiveOverlayData } from "@/features/reconnect-live-transport/model";
import { PlayerDeckPanel } from "@/widgets/overlay-player-deck/player-deck-panel";
import { OpponentCardsPanel } from "@/widgets/overlay-opponent-cards/opponent-cards-panel";

export function LiveOverlayWindowPage() {
  const { bridge } = useServices();
  const { snapshot } = useLiveOverlayData();
  const { inspectedCard, inspectCard } = useCardInspectState();
  const [mode, setMode] = useState<OverlayMode>("inspect");
  const [tooltipSide, setTooltipSide] = useState<"left" | "right" | null>(null);

  useEffect(() => {
    let cancelled = false;

    const syncMode = async () => {
      const nextMode = await bridge.getOverlayMode();
      if (!cancelled) {
        setMode((current) => {
          if (current === nextMode) {
            return current;
          }
          return nextMode;
        });
      }
    };

    void syncMode();
    const intervalId = window.setInterval(() => {
      void syncMode();
    }, 600);

    return () => {
      cancelled = true;
      window.clearInterval(intervalId);
    };
  }, [bridge]);

  useEffect(() => {
    if (mode !== "inspect") {
      setTooltipSide(null);
      inspectCard(null);
    }
  }, [inspectCard, mode]);

  const inspectMode = mode === "inspect";
  const isMatchActive = snapshot?.match.status === "active";

  if (!isMatchActive) {
    return <div className="relative h-screen overflow-hidden" />;
  }

  return (
    <div className="relative h-screen overflow-hidden">
      <div className="absolute left-[var(--overlay-left-offset)] top-[var(--overlay-top-offset)]">
        <OpponentCardsPanel
          cards={snapshot?.opponentPlayedCards ?? []}
          inspectMode={inspectMode}
          onInspect={(card) => {
            if (!inspectMode) {
              return;
            }
            inspectCard(card);
            setTooltipSide(card ? "left" : null);
          }}
        />
      </div>
      {inspectMode && tooltipSide === "left" && inspectedCard ? (
        <div className="pointer-events-none absolute left-[calc(var(--overlay-left-offset)+var(--overlay-panel-width)+var(--overlay-panel-gap))] top-[var(--overlay-top-offset)]">
          <CardTooltip card={inspectedCard} />
        </div>
      ) : null}
      <div className="absolute right-[var(--overlay-right-offset)] top-[var(--overlay-top-offset)]">
        <PlayerDeckPanel
          cards={snapshot?.playerDeckRemaining ?? []}
          inspectMode={inspectMode}
          onInspect={(card) => {
            if (!inspectMode) {
              return;
            }
            inspectCard(card);
            setTooltipSide(card ? "right" : null);
          }}
        />
      </div>
      {inspectMode && tooltipSide === "right" && inspectedCard ? (
        <div className="pointer-events-none absolute right-[calc(var(--overlay-right-offset)+var(--overlay-panel-width)+var(--overlay-panel-gap))] top-[var(--overlay-top-offset)]">
          <CardTooltip card={inspectedCard} />
        </div>
      ) : null}
    </div>
  );
}
