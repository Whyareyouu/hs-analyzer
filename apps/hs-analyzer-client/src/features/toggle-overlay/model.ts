import { useEffect, useState } from "react";
import type { OverlayMode } from "@electron/shared/contracts";
import { useServices } from "@/shared/api/services-context";

export function useOverlayModeControl() {
  const { bridge } = useServices();
  const [mode, setMode] = useState<OverlayMode>("inspect");

  useEffect(() => {
    void bridge.getOverlayMode().then(setMode);
  }, [bridge]);

  const toggleMode = async () => {
    const nextMode = mode === "passive" ? "inspect" : "passive";
    const resolved = await bridge.setOverlayMode(nextMode);
    setMode(resolved);
  };

  return { mode, toggleMode };
}
