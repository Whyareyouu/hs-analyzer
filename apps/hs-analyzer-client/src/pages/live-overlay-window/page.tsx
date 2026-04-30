import { useEffect, useState } from "react";
import type { OverlayMode } from "@electron/shared/contracts";
import { useServices } from "@/shared/api/services-context";
import { CardTooltip } from "@/shared/ui/card-tooltip";
import { useCardInspectState } from "@/features/inspect-card/model";
import { useLiveOverlayData } from "@/features/reconnect-live-transport/model";
import { PlayerDeckPanel } from "@/widgets/overlay-player-deck/player-deck-panel";
import { OpponentCardsPanel } from "@/widgets/overlay-opponent-cards/opponent-cards-panel";

const CARD_TOOLTIP_HEIGHT = 350;
const CARD_TOOLTIP_MARGIN = 12;

export function LiveOverlayWindowPage() {
  const { bridge } = useServices();
  const { snapshot } = useLiveOverlayData();
  const { inspectedCard, inspectCard } = useCardInspectState();
  const [mode, setMode] = useState<OverlayMode>("inspect");
  const [tooltipSide, setTooltipSide] = useState<"left" | "right" | null>(null);
  const [tooltipTop, setTooltipTop] = useState<number | null>(null);

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
      setTooltipTop(null);
      inspectCard(null);
    }
  }, [inspectCard, mode]);

  useEffect(() => {
    if (mode !== "inspect") {
      void bridge.setOverlayInspectInteractivity(false);
      return;
    }

    let lastInteractive = false;

    const syncInteractivity = (interactive: boolean) => {
      if (interactive === lastInteractive) {
        return;
      }
      lastInteractive = interactive;
      void bridge.setOverlayInspectInteractivity(interactive);
    };

    const handlePointerMove = (event: PointerEvent) => {
      const target =
        typeof document.elementFromPoint === "function"
          ? document.elementFromPoint(event.clientX, event.clientY)
          : null;
      syncInteractivity(Boolean(target?.closest("[data-overlay-interactive='true']")));
    };

    const handlePointerLeave = () => {
      syncInteractivity(false);
    };

    window.addEventListener("pointermove", handlePointerMove);
    window.addEventListener("blur", handlePointerLeave);
    document.addEventListener("mouseleave", handlePointerLeave);
    void bridge.setOverlayInspectInteractivity(false);

    return () => {
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("blur", handlePointerLeave);
      document.removeEventListener("mouseleave", handlePointerLeave);
      void bridge.setOverlayInspectInteractivity(false);
    };
  }, [bridge, mode]);

  const inspectMode = mode === "inspect";
  const isMatchActive = snapshot?.match.status === "active";

  if (!isMatchActive) {
    return <div className="relative h-screen overflow-hidden" />;
  }

  return (
    <div className="relative h-screen overflow-hidden">
      <div data-overlay-interactive="true" className="absolute left-[var(--overlay-left-offset)] top-[var(--overlay-top-offset)]">
        <OpponentCardsPanel
          cards={snapshot?.opponentPlayedCards ?? []}
          inspectMode={inspectMode}
          onInspect={(card, anchorRect) => {
            if (!inspectMode) {
              return;
            }
            inspectCard(card);
            setTooltipSide(card ? "left" : null);
            setTooltipTop(card && anchorRect ? getTooltipTop(anchorRect) : null);
          }}
        />
      </div>
      {inspectMode && tooltipSide === "left" && inspectedCard && tooltipTop !== null ? (
        <div
          data-overlay-interactive="true"
          data-testid="overlay-card-tooltip"
          className="pointer-events-none absolute left-[calc(var(--overlay-left-offset)+var(--overlay-panel-width)+var(--overlay-panel-gap))]"
          style={{ top: `${tooltipTop}px` }}
        >
          <CardTooltip card={inspectedCard} />
        </div>
      ) : null}
      <div data-overlay-interactive="true" className="absolute right-[var(--overlay-right-offset)] top-[var(--overlay-top-offset)]">
        <PlayerDeckPanel
          cards={snapshot?.playerDeckRemaining ?? []}
          otherCards={snapshot?.playerOtherPlayedCards ?? []}
          inspectMode={inspectMode}
          onInspect={(card, anchorRect) => {
            if (!inspectMode) {
              return;
            }
            inspectCard(card);
            setTooltipSide(card ? "right" : null);
            setTooltipTop(card && anchorRect ? getTooltipTop(anchorRect) : null);
          }}
        />
      </div>
      {inspectMode && tooltipSide === "right" && inspectedCard && tooltipTop !== null ? (
        <div
          data-overlay-interactive="true"
          data-testid="overlay-card-tooltip"
          className="pointer-events-none absolute right-[calc(var(--overlay-right-offset)+var(--overlay-panel-width)+var(--overlay-panel-gap))]"
          style={{ top: `${tooltipTop}px` }}
        >
          <CardTooltip card={inspectedCard} />
        </div>
      ) : null}
    </div>
  );
}

function getTooltipTop(anchorRect: DOMRect): number {
  const maxTop = Math.max(CARD_TOOLTIP_MARGIN, window.innerHeight - CARD_TOOLTIP_HEIGHT - CARD_TOOLTIP_MARGIN);
  return Math.min(Math.max(anchorRect.top, CARD_TOOLTIP_MARGIN), maxTop);
}
