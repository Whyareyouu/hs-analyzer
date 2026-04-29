import { parseOverlayEnvelope, parseOverlaySnapshot } from "@/shared/api/decoders";
import type { OverlayEnvelope, OverlaySnapshot } from "@/shared/model/overlay";

export class LiveOverlayClient {
  constructor(private readonly baseUrl: string) {}

  async getSnapshot(): Promise<OverlaySnapshot> {
    const response = await fetch(`${this.baseUrl}/api/v1/live/overlay-snapshot`);
    if (!response.ok) {
      throw new Error(`live snapshot request failed with status ${response.status}`);
    }
    return parseOverlaySnapshot(await response.json());
  }

  connect(onMessage: (message: OverlayEnvelope) => void, onDisconnect?: () => void): () => void {
    const socketUrl = this.baseUrl.replace(/^http/, "ws");
    const socket = new WebSocket(`${socketUrl}/ws/live`);
    socket.addEventListener("message", (event) => {
      const raw = typeof event.data === "string" ? JSON.parse(event.data) : {};
      onMessage(parseOverlayEnvelope(raw));
    });
    socket.addEventListener("close", () => {
      onDisconnect?.();
    });
    socket.addEventListener("error", () => {
      onDisconnect?.();
    });
    return () => {
      socket.close();
    };
  }
}
