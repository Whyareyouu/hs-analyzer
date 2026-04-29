import { useEffect, useState } from "react";
import { useServices } from "@/shared/api/services-context";
import type { OverlaySnapshot } from "@/shared/model/overlay";

export function useLiveOverlayData() {
  const { liveClient } = useServices();
  const [snapshot, setSnapshot] = useState<OverlaySnapshot | null>(null);
  const [status, setStatus] = useState<"connecting" | "connected" | "offline">("connecting");
  const [lastResetMatchId, setLastResetMatchId] = useState("");

  useEffect(() => {
    let unsubscribe: (() => void) | undefined;
    let cancelled = false;
    let reconnectTimer: number | undefined;

    const scheduleReconnect = () => {
      if (cancelled || reconnectTimer) {
        return;
      }
      reconnectTimer = window.setTimeout(() => {
        reconnectTimer = undefined;
        void connect();
      }, 1500);
    };

    const connect = async () => {
      setStatus("connecting");
      try {
        const initial = await liveClient.getSnapshot();
        if (cancelled) {
          return;
        }
        setSnapshot(initial);
        setStatus("connected");
        unsubscribe = liveClient.connect((message) => {
          if (message.type === "overlay_reset") {
            setLastResetMatchId(message.matchId ?? "");
          }
          if (message.snapshot) {
            setSnapshot(message.snapshot);
          }
          if (message.delta) {
            setSnapshot(message.delta);
          }
          if (message.health) {
            setSnapshot((current) => (current ? { ...current, health: message.health ?? current.health } : current));
          }
        }, () => {
          if (!cancelled) {
            setStatus("offline");
            scheduleReconnect();
          }
        });
      } catch {
        if (!cancelled) {
          setStatus("offline");
          scheduleReconnect();
        }
      }
    };

    void connect();

    return () => {
      cancelled = true;
      if (reconnectTimer) {
        window.clearTimeout(reconnectTimer);
      }
      unsubscribe?.();
    };
  }, [liveClient]);

  return { snapshot, status, lastResetMatchId };
}
