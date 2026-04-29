import { contextBridge, ipcRenderer } from "electron";
import { createNativeBridge } from "./bridge";

const bridge = createNativeBridge({
  invoke: (channel: string, payload?: unknown) => ipcRenderer.invoke(channel, payload),
  subscribe: <T>(channel: string, listener: (payload: T) => void) => {
    const handler = (_event: unknown, payload: unknown) => {
      listener(payload as T);
    };
    ipcRenderer.on(channel, handler);
    return () => {
      ipcRenderer.removeListener(channel, handler);
    };
  },
});

contextBridge.exposeInMainWorld("hsAnalyzer", bridge);
