import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AppProviders } from "@/app/providers/app-providers";
import { SettingsPage } from "./page";

const bridgeState = vi.hoisted(() => ({
  mode: "inspect" as "passive" | "inspect",
  setOverlayModeCalls: [] as Array<"passive" | "inspect">,
}));

vi.mock("@/shared/config/app-config", () => ({
  getNativeBridge: () => ({
    getConfig: async () => ({ backendBaseUrl: "http://127.0.0.1:8080", liveBaseUrl: "http://127.0.0.1:8765" }),
    getOverlayMode: async () => bridgeState.mode,
    setOverlayMode: async (mode: "passive" | "inspect") => {
      bridgeState.mode = mode;
      bridgeState.setOverlayModeCalls.push(mode);
      return mode;
    },
    getOverlayDiagnostics: async () => ({
      attached: false,
      degraded: true,
      gameDetected: false,
      reason: "test",
      hearthstoneMode: "unsupported",
      bounds: null,
    }),
    onOverlayDiagnostics: () => () => undefined,
  }),
}));

describe("SettingsPage", () => {
  beforeEach(() => {
    bridgeState.mode = "inspect";
    bridgeState.setOverlayModeCalls = [];
  });

  it("keeps inspect mode enabled by default and lets the user disable it", async () => {
    render(
      <AppProviders>
        <SettingsPage />
      </AppProviders>,
    );

    expect(await screen.findByText("включен")).toBeInTheDocument();
    const switchButton = screen.getByRole("switch", { name: /Отключить inspect mode/i });
    expect(switchButton).toHaveAttribute("aria-checked", "true");

    await userEvent.click(switchButton);

    await waitFor(() => expect(bridgeState.setOverlayModeCalls).toEqual(["passive"]));
    expect(screen.getByText("выключен")).toBeInTheDocument();
    expect(screen.getByRole("switch", { name: /Включить inspect mode/i })).toHaveAttribute("aria-checked", "false");
  });
});
