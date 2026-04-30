import type { NativeBridge, NativeBridgeConfig, OverlayDiagnostics, OverlayMode } from "@electron/shared/contracts";

export interface BridgeDeps {
  invoke<T>(channel: string, payload?: unknown): Promise<T>;
  subscribe<T>(channel: string, listener: (payload: T) => void): () => void;
}

export const BRIDGE_CHANNELS = {
  getConfig: "client:get-config",
  getOverlayMode: "overlay:get-mode",
  setOverlayMode: "overlay:set-mode",
  setOverlayInspectInteractivity: "overlay:set-inspect-interactivity",
  getOverlayDiagnostics: "overlay:get-diagnostics",
  onOverlayDiagnostics: "overlay:diagnostics",
} as const;

export function createNativeBridge(deps: BridgeDeps): NativeBridge {
  return {
    getConfig(): Promise<NativeBridgeConfig> {
      return deps.invoke<NativeBridgeConfig>(BRIDGE_CHANNELS.getConfig);
    },
    getOverlayMode(): Promise<OverlayMode> {
      return deps.invoke<OverlayMode>(BRIDGE_CHANNELS.getOverlayMode);
    },
    setOverlayMode(mode: OverlayMode): Promise<OverlayMode> {
      return deps.invoke<OverlayMode>(BRIDGE_CHANNELS.setOverlayMode, mode);
    },
    setOverlayInspectInteractivity(active: boolean): Promise<void> {
      return deps.invoke<void>(BRIDGE_CHANNELS.setOverlayInspectInteractivity, active);
    },
    getOverlayDiagnostics(): Promise<OverlayDiagnostics> {
      return deps.invoke<OverlayDiagnostics>(BRIDGE_CHANNELS.getOverlayDiagnostics);
    },
    onOverlayDiagnostics(listener: (value: OverlayDiagnostics) => void): () => void {
      return deps.subscribe<OverlayDiagnostics>(BRIDGE_CHANNELS.onOverlayDiagnostics, listener);
    },
  };
}
