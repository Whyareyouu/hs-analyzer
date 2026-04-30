import { describe, expect, it, vi } from "vitest";
import { BRIDGE_CHANNELS, createNativeBridge } from "./bridge";

describe("createNativeBridge", () => {
  it("uses a typed preload surface without exposing arbitrary node api", async () => {
    const invoke = vi.fn(async (_channel: string, payload?: unknown) => payload ?? { liveBaseUrl: "", backendBaseUrl: "" });
    const subscribe = vi.fn(() => () => undefined);

    const bridge = createNativeBridge({ invoke, subscribe });

    await bridge.getConfig();
    await bridge.getOverlayMode();
    await bridge.setOverlayMode("inspect");
    await bridge.setOverlayInspectInteractivity(true);
    await bridge.getOverlayDiagnostics();
    bridge.onOverlayDiagnostics(() => undefined);

    expect(invoke).toHaveBeenCalledWith(BRIDGE_CHANNELS.getConfig);
    expect(invoke).toHaveBeenCalledWith(BRIDGE_CHANNELS.getOverlayMode);
    expect(invoke).toHaveBeenCalledWith(BRIDGE_CHANNELS.setOverlayMode, "inspect");
    expect(invoke).toHaveBeenCalledWith(BRIDGE_CHANNELS.setOverlayInspectInteractivity, true);
    expect(invoke).toHaveBeenCalledWith(BRIDGE_CHANNELS.getOverlayDiagnostics);
    expect(subscribe).toHaveBeenCalledTimes(1);
  });
});
