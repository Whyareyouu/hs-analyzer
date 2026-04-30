import { app, BrowserWindow, globalShortcut, ipcMain, screen } from "electron";
import { execFile } from "node:child_process";
import fs from "node:fs";
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
const electronDataRoot = path.join(projectRoot, ".tmp", "electron");
const preloadPath = path.join(__dirname, "..", "preload", "index.js");
const execFileAsync = promisify(execFile);

fs.mkdirSync(electronDataRoot, { recursive: true });
app.setPath("userData", path.join(electronDataRoot, "user-data"));
app.setPath("sessionData", path.join(electronDataRoot, "session-data"));

interface PowerShellWindowResult {
  detected: boolean;
  active?: boolean;
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
using System.Text;
using System.Runtime.InteropServices;
public static class HsAnalyzerUser32 {
  public delegate bool EnumWindowsProc(IntPtr hWnd, IntPtr lParam);
  [StructLayout(LayoutKind.Sequential)]
  public struct RECT {
    public int Left;
    public int Top;
    public int Right;
    public int Bottom;
  }
  [DllImport("user32.dll")]
  public static extern bool EnumWindows(EnumWindowsProc callback, IntPtr lParam);
  [DllImport("user32.dll")]
  public static extern bool IsWindowVisible(IntPtr hWnd);
  [DllImport("user32.dll", CharSet = CharSet.Unicode)]
  public static extern int GetWindowText(IntPtr hWnd, StringBuilder lpString, int nMaxCount);
  [DllImport("user32.dll", CharSet = CharSet.Unicode)]
  public static extern int GetWindowTextLength(IntPtr hWnd);
  [DllImport("user32.dll")]
  public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out int processId);
  [DllImport("user32.dll")]
  public static extern bool GetWindowRect(IntPtr hWnd, out RECT rect);
  [DllImport("user32.dll")]
  public static extern IntPtr GetForegroundWindow();
}
"@
$foregroundWindow = [HsAnalyzerUser32]::GetForegroundWindow()
$foregroundProcessId = 0
if ($foregroundWindow -ne [IntPtr]::Zero) {
  [void][HsAnalyzerUser32]::GetWindowThreadProcessId($foregroundWindow, [ref]$foregroundProcessId)
}
$allProcesses = @(Get-Process -ErrorAction SilentlyContinue | Where-Object { $_.ProcessName -like 'Hearthstone*' })
if ($allProcesses.Count -eq 0) {
  $windowCandidates = New-Object System.Collections.Generic.List[object]
  $enumWithoutProcess = [HsAnalyzerUser32+EnumWindowsProc]{
    param([IntPtr]$hWnd, [IntPtr]$lParam)
    if (-not [HsAnalyzerUser32]::IsWindowVisible($hWnd)) {
      return $true
    }
    $length = [HsAnalyzerUser32]::GetWindowTextLength($hWnd)
    if ($length -le 0) {
      return $true
    }
    $builder = New-Object System.Text.StringBuilder ($length + 1)
    [void][HsAnalyzerUser32]::GetWindowText($hWnd, $builder, $builder.Capacity)
    $title = $builder.ToString().Trim()
    if ($title -notlike '*Hearthstone*') {
      return $true
    }
    $rect = New-Object HsAnalyzerUser32+RECT
    if (-not [HsAnalyzerUser32]::GetWindowRect($hWnd, [ref]$rect)) {
      return $true
    }
    $width = [Math]::Max(0, $rect.Right - $rect.Left)
    $height = [Math]::Max(0, $rect.Bottom - $rect.Top)
    if ($width -lt 200 -or $height -lt 200) {
      return $true
    }
    $windowCandidates.Add([pscustomobject]@{
      title = $title
      process_name = ''
      process_id = 0
      x = $rect.Left
      y = $rect.Top
      width = $width
      height = $height
      area = $width * $height
      is_foreground = $hWnd -eq $foregroundWindow
      is_active_process = $false
    }) | Out-Null
    return $true
  }
  [void][HsAnalyzerUser32]::EnumWindows($enumWithoutProcess, [IntPtr]::Zero)
  if ($windowCandidates.Count -eq 0) {
  [Console]::Out.Write('{"detected":false,"reason":"Hearthstone process not found","bounds":null}')
  exit 0
}
  $window = ($windowCandidates | Where-Object { $_.is_foreground } | Select-Object -First 1)
  if ($null -eq $window) {
    $window = $windowCandidates | Sort-Object area -Descending | Select-Object -First 1
  }
  $payload = [ordered]@{
    detected = $true
    title = $window.title
    active = [bool]$window.is_foreground
    reason = if ($window.is_foreground) { "Hearthstone window detected by title match" } else { "Hearthstone window detected by title match but is not active" }
    bounds = @{
      x = $window.x
      y = $window.y
      width = $window.width
      height = $window.height
    }
  }
  $payload | ConvertTo-Json -Compress
  exit 0
}
$processById = @{}
foreach ($process in $allProcesses) {
  $processById[[int]$process.Id] = $process
}
$windowCandidates = New-Object System.Collections.Generic.List[object]
$enumWindows = [HsAnalyzerUser32+EnumWindowsProc]{
  param([IntPtr]$hWnd, [IntPtr]$lParam)
  if (-not [HsAnalyzerUser32]::IsWindowVisible($hWnd)) {
    return $true
  }
  $windowPid = 0
  [void][HsAnalyzerUser32]::GetWindowThreadProcessId($hWnd, [ref]$windowPid)
  $belongsToProcess = $processById.ContainsKey([int]$windowPid)
  $length = [HsAnalyzerUser32]::GetWindowTextLength($hWnd)
  $builder = New-Object System.Text.StringBuilder ($length + 1)
  [void][HsAnalyzerUser32]::GetWindowText($hWnd, $builder, $builder.Capacity)
  $title = $builder.ToString().Trim()
  $matchesTitle = $title -like '*Hearthstone*'
  if (-not $belongsToProcess -and -not $matchesTitle) {
    return $true
  }
  $rect = New-Object HsAnalyzerUser32+RECT
  if (-not [HsAnalyzerUser32]::GetWindowRect($hWnd, [ref]$rect)) {
    return $true
  }
  $width = [Math]::Max(0, $rect.Right - $rect.Left)
  $height = [Math]::Max(0, $rect.Bottom - $rect.Top)
  if ($width -lt 200 -or $height -lt 200) {
    return $true
  }
  $processName = if ($belongsToProcess) { $processById[[int]$windowPid].ProcessName } else { '' }
  $windowCandidates.Add([pscustomobject]@{
    title = $title
    process_name = $processName
    process_id = [int]$windowPid
    x = $rect.Left
    y = $rect.Top
    width = $width
    height = $height
    area = $width * $height
    is_foreground = $hWnd -eq $foregroundWindow
    is_active_process = [int]$windowPid -eq [int]$foregroundProcessId
  }) | Out-Null
  return $true
}
[void][HsAnalyzerUser32]::EnumWindows($enumWindows, [IntPtr]::Zero)
if ($windowCandidates.Count -eq 0) {
  [Console]::Out.Write('{"detected":true,"reason":"Hearthstone process found, but no visible top-level window was detected","bounds":null}')
  exit 0
}
$window = ($windowCandidates | Where-Object { $_.is_foreground -or $_.is_active_process } | Sort-Object area -Descending | Select-Object -First 1)
if ($null -eq $window) {
  $window = $windowCandidates | Sort-Object area -Descending | Select-Object -First 1
}
$isActive = [bool]($window.is_foreground -or $window.is_active_process)
$payload = [ordered]@{
  detected = $true
  title = $window.title
  active = $isActive
  reason = if ($isActive) { "Hearthstone window detected" } else { "Hearthstone window detected but is not active" }
  bounds = @{
    x = $window.x
    y = $window.y
    width = $window.width
    height = $window.height
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
        active: parsed.active === true,
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
  configureOverlayWindow(overlayWindow);
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
  ipcMain.handle(BRIDGE_CHANNELS.getOverlayMode, () => overlayController?.getMode() ?? "inspect");
  ipcMain.handle(BRIDGE_CHANNELS.setOverlayMode, (_event, mode: OverlayMode) => {
    const nextMode = overlayController?.setMode(mode) ?? "inspect";
    mainWindow?.webContents.send(BRIDGE_CHANNELS.onOverlayDiagnostics, overlayController?.getDiagnostics());
    overlayWindow?.webContents.send(BRIDGE_CHANNELS.onOverlayDiagnostics, overlayController?.getDiagnostics());
    return nextMode;
  });
  ipcMain.handle(BRIDGE_CHANNELS.setOverlayInspectInteractivity, (_event, active: boolean) => {
    overlayController?.setInspectInteractivity(active);
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
    const currentMode = overlayController?.getMode() ?? "inspect";
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
    active: false,
    supportedMode: "unsupported" as const,
    bounds: null,
    reason,
	};
}

function detectedButUnsupportedWindowState(reason: string) {
  return {
    detected: true,
    active: false,
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

function configureOverlayWindow(window: BrowserWindow) {
  window.setTitle("");
  window.setMenuBarVisibility(false);
  window.setAlwaysOnTop(true, "screen-saver", 1);
  window.setVisibleOnAllWorkspaces(true, { visibleOnFullScreen: true });
  window.setSkipTaskbar(true);
  window.moveTop();
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
