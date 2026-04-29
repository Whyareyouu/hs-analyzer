import type { NativeBridge } from "@electron/shared/contracts";

declare global {
  interface Window {
    hsAnalyzer?: NativeBridge;
  }
}

export {};
