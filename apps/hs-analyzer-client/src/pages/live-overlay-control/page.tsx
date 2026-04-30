import { useEffect, useState } from "react";
import type { OverlayDiagnostics } from "@electron/shared/contracts";
import { useServices } from "@/shared/api/services-context";
import { useCardInspectState } from "@/features/inspect-card/model";
import { useLiveOverlayData } from "@/features/reconnect-live-transport/model";
import { useOverlayModeControl } from "@/features/toggle-overlay/model";
import { Button } from "@/shared/ui/button";
import { CardTooltip } from "@/shared/ui/card-tooltip";
import { Pill } from "@/shared/ui/pill";
import { PlayerDeckPanel } from "@/widgets/overlay-player-deck/player-deck-panel";
import { OpponentCardsPanel } from "@/widgets/overlay-opponent-cards/opponent-cards-panel";

const defaultDiagnostics: OverlayDiagnostics = {
  attached: false,
  degraded: true,
  gameDetected: false,
  reason: "Overlay diagnostics not loaded yet",
  hearthstoneMode: "unsupported",
  bounds: null,
};

export function LiveOverlayControlPage() {
  const { bridge } = useServices();
  const { mode, toggleMode } = useOverlayModeControl();
  const { snapshot, status, lastResetMatchId } = useLiveOverlayData();
  const { inspectedCard, inspectCard } = useCardInspectState();
  const [diagnostics, setDiagnostics] = useState<OverlayDiagnostics>(defaultDiagnostics);

  useEffect(() => {
    let dispose = () => undefined;
    void bridge.getOverlayDiagnostics().then(setDiagnostics);
    dispose = bridge.onOverlayDiagnostics(setDiagnostics);
    return () => {
      dispose();
    };
  }, [bridge]);

  const inspectMode = mode === "inspect";

  return (
    <div className="min-h-screen p-6">
      <div className="mb-6 flex flex-wrap items-center gap-3">
        <Pill tone={status === "connected" ? "success" : "warning"}>{status}</Pill>
        <Pill tone={diagnostics.degraded ? "warning" : "success"}>{diagnostics.hearthstoneMode}</Pill>
        {lastResetMatchId ? <Pill tone="info">reset from {lastResetMatchId}</Pill> : null}
        <Button onClick={() => void toggleMode()}>{inspectMode ? "Выключить inspect mode" : "Включить inspect mode"}</Button>
      </div>
      <div className="mb-4 rounded-xl border border-app-border/70 bg-app-panel/50 p-4 text-sm text-app-muted">
        {diagnostics.reason}. Overlay {diagnostics.attached ? "привязан к окну игры" : "в standby"}.
      </div>
      <div className="grid gap-6 xl:grid-cols-[1fr_auto]">
        <div className="flex items-start justify-between gap-[var(--overlay-panel-gap)]">
          <PlayerDeckPanel cards={snapshot?.playerDeckRemaining ?? []} otherCards={snapshot?.playerOtherPlayedCards ?? []} inspectMode={inspectMode} onInspect={inspectCard} />
          <OpponentCardsPanel cards={snapshot?.opponentPlayedCards ?? []} inspectMode={inspectMode} onInspect={inspectCard} />
        </div>
        <div className="sticky top-6">
          <CardTooltip card={inspectMode ? inspectedCard : null} />
        </div>
      </div>
    </div>
  );
}
