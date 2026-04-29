import { describe, expect, it, vi } from "vitest";
import { OverlayWindowController, type GameWindowLocator, type ManagedWindow } from "./overlay-controller";

function createWindowMock(): ManagedWindow {
  return {
    showInactive: vi.fn(),
    hide: vi.fn(),
    setBounds: vi.fn(),
    setIgnoreMouseEvents: vi.fn(),
    isVisible: vi.fn(() => true),
  };
}

describe("OverlayWindowController", () => {
  it("keeps overlay passive by default and follows supported Hearthstone bounds", async () => {
    const windowMock = createWindowMock();
    const locator: GameWindowLocator = {
      locate: vi.fn(async () => ({
        detected: true,
        supportedMode: "windowed",
        bounds: { x: 10, y: 20, width: 1440, height: 900 },
        reason: "ok",
      })),
    };

    const controller = new OverlayWindowController(windowMock, locator);
    const diagnostics = await controller.syncBounds();

    expect(diagnostics.attached).toBe(true);
    expect(windowMock.setBounds).toHaveBeenCalledWith({ x: 10, y: 20, width: 1440, height: 900 });
    expect(windowMock.setIgnoreMouseEvents).toHaveBeenCalledWith(true, { forward: true });
  });

  it("switches between passive and inspect modes", () => {
    const windowMock = createWindowMock();
    const controller = new OverlayWindowController(windowMock, {
      locate: async () => ({
        detected: true,
        supportedMode: "borderless",
        bounds: { x: 0, y: 0, width: 100, height: 100 },
        reason: "ok",
      }),
    });

    controller.setMode("inspect");
    expect(controller.getMode()).toBe("inspect");
  });

  it("hides overlay for unsupported fullscreen mode", async () => {
    const windowMock = createWindowMock();
    const controller = new OverlayWindowController(windowMock, {
      locate: async () => ({
        detected: false,
        supportedMode: "unsupported",
        bounds: null,
        reason: "exclusive fullscreen is not supported",
      }),
    });

    const diagnostics = await controller.syncBounds();
    expect(diagnostics.degraded).toBe(true);
    expect(windowMock.hide).toHaveBeenCalled();
  });

  it("shows detached preview when Hearthstone is detected but no attachable window is available", async () => {
    const windowMock = createWindowMock();
    const controller = new OverlayWindowController(
      windowMock,
      {
        locate: async () => ({
          detected: true,
          supportedMode: "unsupported",
          bounds: null,
          reason: "Hearthstone process exists but does not expose a desktop window",
        }),
      },
      () => ({ x: 20, y: 20, width: 680, height: 900 }),
    );

    const diagnostics = await controller.syncBounds();
    expect(diagnostics.attached).toBe(false);
    expect(diagnostics.bounds).toEqual({ x: 20, y: 20, width: 680, height: 900 });
    expect(windowMock.showInactive).toHaveBeenCalled();
  });

  it("keeps fullscreen fallback visible when Hearthstone detection fails completely", async () => {
    const windowMock = createWindowMock();
    const controller = new OverlayWindowController(
      windowMock,
      {
        locate: async () => ({
          detected: false,
          supportedMode: "unsupported",
          bounds: null,
          reason: "Hearthstone process not found",
        }),
      },
      () => ({ x: 0, y: 0, width: 1920, height: 1080 }),
    );

    const diagnostics = await controller.syncBounds();
    expect(diagnostics.attached).toBe(false);
    expect(diagnostics.degraded).toBe(true);
    expect(diagnostics.bounds).toEqual({ x: 0, y: 0, width: 1920, height: 1080 });
    expect(windowMock.showInactive).toHaveBeenCalled();
  });
});
