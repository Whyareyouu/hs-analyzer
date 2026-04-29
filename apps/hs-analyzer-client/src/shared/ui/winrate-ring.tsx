import { formatWinrate } from "@/shared/lib/format";

export function WinrateRing({ value, label }: { value: number; label: string }) {
  const normalized = Math.max(0, Math.min(1, value));
  const angle = normalized * 360;

  return (
    <div className="flex flex-col items-center gap-3">
      <div
        className="grid place-items-center rounded-full border border-app-border/80 bg-app-panel p-3 shadow-glow"
        style={{
          width: "var(--dashboard-winrate-ring-size)",
          height: "var(--dashboard-winrate-ring-size)",
          backgroundImage: `conic-gradient(hsl(var(--color-accent-primary)) ${angle}deg, rgba(255,255,255,0.07) ${angle}deg 360deg)`,
        }}
      >
        <div className="grid h-[70%] w-[70%] place-items-center rounded-full bg-[#091120] text-center">
          <div>
            <div className="text-4xl font-bold text-app-text">{formatWinrate(normalized)}</div>
            <div className="text-xs uppercase tracking-[0.2em] text-app-muted">{label}</div>
          </div>
        </div>
      </div>
    </div>
  );
}
