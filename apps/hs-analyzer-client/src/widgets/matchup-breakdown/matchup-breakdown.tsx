import { useState } from "react";
import type { MatchupEntry } from "@/shared/model/dashboard";
import { getMetricTone, formatWinrate } from "@/shared/lib/format";
import { Panel } from "@/shared/ui/panel";
import { Pill } from "@/shared/ui/pill";

export function MatchupBreakdown({ matchups }: { matchups: MatchupEntry[] }) {
  const [open, setOpen] = useState(false);

  return (
    <Panel className="space-y-4">
      <button
        type="button"
        aria-expanded={open}
        onClick={() => setOpen((current) => !current)}
        className="flex w-full items-center justify-between text-left"
      >
        <div>
          <div className="text-xs uppercase tracking-[0.2em] text-app-muted">Analytics</div>
          <div className="text-xl font-semibold text-app-text">Матчапы</div>
        </div>
        <div className="flex items-center gap-3">
          <Pill>{matchups.length}</Pill>
          <span className="text-xs uppercase tracking-[0.18em] text-app-muted">{open ? "Hide" : "Show"}</span>
        </div>
      </button>
      {!open ? null : <div className="space-y-3">
        {matchups.map((matchup) => {
          const tone = getMetricTone(matchup.winrate);
          return (
            <div key={`${matchup.opponentClass}-${matchup.opponentArchetype}`} className="grid gap-2 rounded-lg border border-app-border/50 bg-app-panelStrong/60 p-4">
              <div className="flex items-center justify-between">
                <div className="text-sm font-medium text-app-text">{matchup.opponentArchetype || matchup.opponentClass}</div>
                <Pill tone={tone}>{formatWinrate(matchup.winrate)}</Pill>
              </div>
              <div className="h-2 overflow-hidden rounded-full bg-app-panel">
                <div className="h-full rounded-full bg-app-accent" style={{ width: `${Math.round(matchup.winrate * 100)}%` }} />
              </div>
            </div>
          );
        })}
      </div>}
    </Panel>
  );
}
