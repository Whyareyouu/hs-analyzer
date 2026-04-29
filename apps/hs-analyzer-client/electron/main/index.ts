import { app, BrowserWindow, globalShortcut, ipcMain, screen } from "electron";
import { execFile } from "node:child_process";
import path from "node:path";
import { promisify } from "node:util";
import { fileURLToPath } from "node:url";
import type { OverlayBounds } from "@electron/shared/contracts";
import type { NativeBridgeConfig, OverlayMode } from "@electron/shared/contracts";
import { createMainWindowConfig, createOverlayWindowConfig } from "./windows.js";
import { OverlayWindowController, type GameWindowLocator } from "./overlay-controller.js";
import { BRIDGE_CHANNELS } from "../preload/bridge.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const projectRoot = path.resolve(__dirname, "..", "..");
const preloadPath = path.join(__dirname, "..", "preload", "index.js");
const execFileAsync = promisify(execFile);

interface PowerShellWindowResult {
  detected: boolean;
  title?: string;
  reason: string;
  bounds: OverlayBounds | null;
}

class WindowsHearthstoneLocator implements GameWindowLocator {
  async locate() {
    if (process.platform !== "win32") {
      return unsupportedWindowState("Overlay attachment is currently implemented only for Windows");
    }

    try {
      const script = `
$ErrorActionPreference = 'Stop'
Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
public static class HsAnalyzerUser32 {
  [StructLayout(LayoutKind.Sequential)]
  public struct RECT {
    public int Left;
    public int Top;
    public int Right;
    public int Bottom;
  }
  [DllImport("user32.dll")]
  public static extern bool GetWindowRect(IntPtr hWnd, out RECT rect);
}
"@
$allProcesses = @(Get-Process Hearthstone -ErrorAction SilentlyContinue)
if ($allProcesses.Count -eq 0) {
  [Console]::Out.Write('{"detected":false,"reason":"Hearthstone process not found","bounds":null}')
  exit 0
}
$process = $allProcesses |
  Where-Object { $_.MainWindowHandle -ne 0 } |
  Select-Object -First 1
if (-not $process) {
  [Console]::Out.Write('{"detected":true,"reason":"Hearthstone process found, but it does not expose an attachable desktop window","bounds":null}')
  exit 0
}
$rect = New-Object HsAnalyzerUser32+RECT
if (-not [HsAnalyzerUser32]::GetWindowRect([IntPtr]$process.MainWindowHandle, [ref]$rect)) {
  [Console]::Out.Write('{"detected":true,"reason":"Hearthstone process found, but its window bounds are unavailable","bounds":null}')
  exit 0
}
$payload = [ordered]@{
  detected = $true
  title = $process.MainWindowTitle
  reason = "Hearthstone window detected"
  bounds = @{
    x = $rect.Left
    y = $rect.Top
    width = [Math]::Max(0, $rect.Right - $rect.Left)
    height = [Math]::Max(0, $rect.Bottom - $rect.Top)
  }
}
$payload | ConvertTo-Json -Compress
`;
      const { stdout } = await execFileAsync("powershell.exe", ["-NoProfile", "-Command", script], { windowsHide: true });
      const parsed = JSON.parse(stdout.trim()) as PowerShellWindowResult;
      if (!parsed.detected || !parsed.bounds || parsed.bounds.width <= 0 || parsed.bounds.height <= 0) {
        if (parsed.detected) {
          return detectedButUnsupportedWindowState(parsed.reason || "Hearthstone window is unavailable");
        }
        return unsupportedWindowState(parsed.reason || "Hearthstone window is unavailable");
      }
      return {
        detected: true,
        supportedMode: classifyWindowMode(parsed.bounds),
        bounds: parsed.bounds,
        reason: parsed.reason || parsed.title || "Hearthstone window detected",
      };
    } catch (error) {
      return unsupportedWindowState(`Failed to inspect Hearthstone window: ${formatError(error)}`);
    }
  }
}

let mainWindow: BrowserWindow | null = null;
let overlayWindow: BrowserWindow | null = null;
let overlayController: OverlayWindowController | null = null;
let overlaySyncTimer: NodeJS.Timeout | null = null;

const bridgeConfig: NativeBridgeConfig = {
  liveBaseUrl: "http://127.0.0.1:8765",
  backendBaseUrl: "http://127.0.0.1:8080",
};

async function createWindows() {
  mainWindow = new BrowserWindow(createMainWindowConfig(preloadPath));
  overlayWindow = new BrowserWindow(createOverlayWindowConfig(preloadPath));
  overlayController = new OverlayWindowController(overlayWindow, new WindowsHearthstoneLocator(), buildPreviewBounds);

  const rendererEntry = path.join(projectRoot, "dist", "index.html");
  await mainWindow.loadFile(rendererEntry, { hash: "/" });
  await overlayWindow.loadFile(rendererEntry, { hash: "/overlay-window" });
  overlayWindow.hide();

  registerBridge();
  registerShortcuts();
  const syncOverlay = async () => {
    if (!overlayController) {
      return;
    }
    await overlayController.syncBounds();
    broadcastOverlayDiagnostics();
  };
  void syncOverlay();
  overlaySyncTimer = setInterval(() => {
    void syncOverlay();
  }, 1500);
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

function broadcastOverlayDiagnostics() {
  const diagnostics = overlayController?.getDiagnostics();
  if (!diagnostics) {
    return;
  }
  mainWindow?.webContents.send(BRIDGE_CHANNELS.onOverlayDiagnostics, diagnostics);
  overlayWindow?.webContents.send(BRIDGE_CHANNELS.onOverlayDiagnostics, diagnostics);
}

function registerShortcuts() {
  globalShortcut.register("CommandOrControl+Shift+I", () => {
    const currentMode = overlayController?.getMode() ?? "passive";
    overlayController?.setMode(currentMode === "passive" ? "inspect" : "passive");
    broadcastOverlayDiagnostics();
  });
}

function classifyWindowMode(bounds: OverlayBounds) {
  if (bounds.width <= 0 || bounds.height <= 0) {
    return "unsupported" as const;
  }
  return "windowed" as const;
}

function unsupportedWindowState(reason: string) {
	return {
		detected: false,
    supportedMode: "unsupported" as const,
    bounds: null,
    reason,
	};
}

function detectedButUnsupportedWindowState(reason: string) {
  return {
    detected: true,
    supportedMode: "unsupported" as const,
    bounds: null,
    reason,
  };
}

function formatError(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}

function buildPreviewBounds(): OverlayBounds {
  const { bounds } = screen.getPrimaryDisplay();
  return {
    x: bounds.x,
    y: bounds.y,
    width: bounds.width,
    height: bounds.height,
  };
}

app.whenReady().then(createWindows);

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});

app.on("will-quit", () => {
  if (overlaySyncTimer) {
    clearInterval(overlaySyncTimer);
    overlaySyncTimer = null;
  }
  globalShortcut.unregisterAll();
});
