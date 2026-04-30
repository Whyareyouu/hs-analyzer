import type { OverlayBounds, OverlayDiagnostics, OverlayMode } from "@electron/shared/contracts";

export interface ManagedWindow {
  showInactive(): void;
  hide(): void;
  setBounds(bounds: OverlayBounds): void;
  setIgnoreMouseEvents(ignore: boolean, options?: { forward: boolean }): void;
  isVisible(): boolean;
  moveTop?(): void;
}

export interface GameWindowState {
  detected: boolean;
  active: boolean;
  supportedMode: "windowed" | "borderless" | "unsupported";
  bounds: OverlayBounds | null;
  reason: string;
}

export interface GameWindowLocator {
  locate(): Promise<GameWindowState>;
}

export class OverlayWindowController {
  private mode: OverlayMode = "passive";
  private diagnostics: OverlayDiagnostics = {
    attached: false,
    degraded: true,
    gameDetected: false,
    reason: "Overlay is waiting for Hearthstone window",
    hearthstoneMode: "unsupported",
    bounds: null,
  };

  constructor(
    private readonly overlayWindow: ManagedWindow,
    private readonly locator: GameWindowLocator,
    private readonly previewBoundsProvider?: () => OverlayBounds,
  ) {}

  async syncBounds(): Promise<OverlayDiagnostics> {
    const gameState = await this.locator.locate();
    const supported =
      gameState.detected && gameState.active && gameState.supportedMode !== "unsupported" && gameState.bounds !== null;

    this.diagnostics = {
      attached: supported,
      degraded: !supported,
      gameDetected: gameState.detected,
      reason: gameState.reason,
      hearthstoneMode: gameState.supportedMode,
      bounds: supported ? gameState.bounds : null,
    };

    if (supported && gameState.bounds) {
      this.overlayWindow.setBounds(gameState.bounds);
      this.overlayWindow.showInactive();
      this.overlayWindow.moveTop?.();
      this.applyMode();
      return this.diagnostics;
    }

    this.overlayWindow.hide();
    return this.diagnostics;
  }

  setMode(mode: OverlayMode): OverlayMode {
    this.mode = mode;
    this.applyMode();
    return this.mode;
  }

  getMode(): OverlayMode {
    return this.mode;
  }

  getDiagnostics(): OverlayDiagnostics {
    return this.diagnostics;
  }

  private applyMode(): void {
    if (!this.overlayWindow.isVisible()) {
      return;
    }

    if (this.mode === "passive") {
      this.overlayWindow.setIgnoreMouseEvents(true, { forward: true });
      return;
    }

    this.overlayWindow.setIgnoreMouseEvents(false);
  }
}
