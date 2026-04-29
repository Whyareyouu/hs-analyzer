export interface BrowserWindowConfig {
  width: number;
  height: number;
  minWidth?: number;
  minHeight?: number;
  transparent?: boolean;
  frame?: boolean;
  titleBarStyle?: "default" | "hidden" | "hiddenInset" | "customButtonsOnHover";
  titleBarOverlay?: boolean;
  autoHideMenuBar?: boolean;
  alwaysOnTop?: boolean;
  skipTaskbar?: boolean;
  resizable?: boolean;
  hasShadow?: boolean;
  backgroundColor?: string;
  title?: string;
  webPreferences: {
    preload: string;
    contextIsolation: boolean;
    nodeIntegration: boolean;
    sandbox: boolean;
  };
}

export function createMainWindowConfig(preloadPath: string): BrowserWindowConfig {
  return {
    width: 1480,
    height: 920,
    minWidth: 1120,
    minHeight: 760,
    backgroundColor: "#0B1020",
    title: "hs-analyzer-client",
    webPreferences: {
      preload: preloadPath,
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  };
}

export function createOverlayWindowConfig(preloadPath: string): BrowserWindowConfig {
  return {
    width: 1400,
    height: 900,
    transparent: true,
    frame: false,
    title: "",
    titleBarStyle: "hidden",
    titleBarOverlay: false,
    autoHideMenuBar: true,
    alwaysOnTop: true,
    skipTaskbar: true,
    resizable: false,
    hasShadow: false,
    backgroundColor: "#00000000",
    webPreferences: {
      preload: preloadPath,
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  };
}
