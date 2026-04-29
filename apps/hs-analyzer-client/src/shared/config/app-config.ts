import type { NativeBridge, NativeBridgeConfig } from "@electron/shared/contracts";

const fallbackConfig: NativeBridgeConfig = {
  backendBaseUrl: "http://127.0.0.1:8080",
  liveBaseUrl: "http://127.0.0.1:8765",
};

const fallbackBridge: NativeBridge = {
  async getConfig() {
    return fallbackConfig;
  },
  async getOverlayMode() {
    return "passive";
  },
  async setOverlayMode(mode) {
    return mode;
  },
  async getOverlayDiagnostics() {
    return {
      attached: false,
      degraded: true,
      gameDetected: false,
      reason: "Electron bridge is not attached",
      hearthstoneMode: "unsupported",
      bounds: null,
    };
  },
  onOverlayDiagnostics() {
    return () => undefined;
  },
};

export function getNativeBridge(): NativeBridge {
  return window.hsAnalyzer ?? fallbackBridge;
}
