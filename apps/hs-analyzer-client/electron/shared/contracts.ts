export type OverlayMode = "passive" | "inspect";
export type HearthstoneWindowMode = "windowed" | "borderless" | "unsupported";

export interface OverlayBounds {
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface OverlayDiagnostics {
  attached: boolean;
  degraded: boolean;
  gameDetected: boolean;
  reason: string;
  hearthstoneMode: HearthstoneWindowMode;
  bounds: OverlayBounds | null;
}

export interface NativeBridgeConfig {
  backendBaseUrl: string;
  liveBaseUrl: string;
}

export interface NativeBridge {
  getConfig(): Promise<NativeBridgeConfig>;
  getOverlayMode(): Promise<OverlayMode>;
  setOverlayMode(mode: OverlayMode): Promise<OverlayMode>;
  getOverlayDiagnostics(): Promise<OverlayDiagnostics>;
  onOverlayDiagnostics(listener: (value: OverlayDiagnostics) => void): () => void;
}
