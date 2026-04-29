import { app, BrowserWindow, globalShortcut, ipcMain } from "electron";
import path from "node:path";
import { fileURLToPath } from "node:url";
import type { NativeBridgeConfig, OverlayMode } from "@electron/shared/contracts";
import { createMainWindowConfig, createOverlayWindowConfig } from "./windows";
import { OverlayWindowController, type GameWindowLocator } from "./overlay-controller";
import { BRIDGE_CHANNELS } from "../preload/bridge";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const projectRoot = path.resolve(__dirname, "..", "..");
const preloadPath = path.join(__dirname, "..", "preload", "index.js");

class NullGameWindowLocator implements GameWindowLocator {
  async locate() {
    return {
      detected: false,
      supportedMode: "unsupported" as const,
      bounds: null,
      reason: "Hearthstone window locator is not configured in this build",
    };
  }
}

let mainWindow: BrowserWindow | null = null;
let overlayWindow: BrowserWindow | null = null;
let overlayController: OverlayWindowController | null = null;

const bridgeConfig: NativeBridgeConfig = {
  liveBaseUrl: "http://127.0.0.1:8765",
  backendBaseUrl: "http://127.0.0.1:8080",
};

async function createWindows() {
  mainWindow = new BrowserWindow(createMainWindowConfig(preloadPath));
  overlayWindow = new BrowserWindow(createOverlayWindowConfig(preloadPath));
  overlayController = new OverlayWindowController(overlayWindow, new NullGameWindowLocator());

  const rendererEntry = path.join(projectRoot, "dist", "index.html");
  await mainWindow.loadFile(rendererEntry, { hash: "/" });
  await overlayWindow.loadFile(rendererEntry, { hash: "/overlay" });
  overlayWindow.hide();

  registerBridge();
  registerShortcuts();
  void overlayController.syncBounds();
}

function registerBridge() {
  ipcMain.handle(BRIDGE_CHANNELS.getConfig, () => bridgeConfig);
  ipcMain.handle(BRIDGE_CHANNELS.getOverlayMode, () => overlayController?.getMode() ?? "passive");
  ipcMain.handle(BRIDGE_CHANNELS.setOverlayMode, (_event, mode: OverlayMode) => {
    const nextMode = overlayController?.setMode(mode) ?? "passive";
    mainWindow?.webContents.send(BRIDGE_CHANNELS.onOverlayDiagnostics, overlayController?.getDiagnostics());
    overlayWindow?.webContents.send(BRIDGE_CHANNELS.onOverlayDiagnostics, overlayController?.getDiagnostics());
    return nextMode;
  });
  ipcMain.handle(BRIDGE_CHANNELS.getOverlayDiagnostics, () => overlayController?.getDiagnostics());
}

function registerShortcuts() {
  globalShortcut.register("CommandOrControl+Shift+I", () => {
    const currentMode = overlayController?.getMode() ?? "passive";
    overlayController?.setMode(currentMode === "passive" ? "inspect" : "passive");
  });
}

app.whenReady().then(createWindows);

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});

app.on("will-quit", () => {
  globalShortcut.unregisterAll();
});
