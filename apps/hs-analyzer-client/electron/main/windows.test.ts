import { describe, expect, it } from "vitest";
import { createMainWindowConfig, createOverlayWindowConfig } from "./windows";

describe("window config", () => {
  it("enforces electron security boundary", () => {
    const mainWindow = createMainWindowConfig("preload.js");
    const overlayWindow = createOverlayWindowConfig("preload.js");

    for (const current of [mainWindow, overlayWindow]) {
      expect(current.webPreferences.contextIsolation).toBe(true);
      expect(current.webPreferences.nodeIntegration).toBe(false);
      expect(current.webPreferences.sandbox).toBe(true);
    }

    expect(overlayWindow.transparent).toBe(true);
    expect(overlayWindow.frame).toBe(false);
    expect(overlayWindow.title).toBe("");
    expect(overlayWindow.titleBarStyle).toBe("hidden");
    expect(overlayWindow.titleBarOverlay).toBe(false);
    expect(overlayWindow.autoHideMenuBar).toBe(true);
    expect(overlayWindow.alwaysOnTop).toBe(true);
  });
});
