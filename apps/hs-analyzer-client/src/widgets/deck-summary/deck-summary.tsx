import type { DeckOverview } from "@/shared/model/dashboard";
import { formatRecord, formatWinrate } from "@/shared/lib/format";
import { Panel } from "@/shared/ui/panel";
import { WinrateRing } from "@/shared/ui/winrate-ring";

export function DeckSummary({ overview, stale }: { overview: DeckOverview; stale: boolean }) {
  return (
    <div className="space-y-6">
      <Panel className="flex items-center gap-5">
        <div className="grid h-20 w-20 place-items-center rounded-2xl bg-app-accent/15 text-3xl font-bold text-app-accent">
          {overview.className.slice(0, 2)}
        </div>
        <div>
          <div className="text-xs uppercase tracking-[0.2em] text-app-muted">Selected Deck</div>
          <div className="text-3xl font-bold text-app-text">{overview.name}</div>
          <div className="mt-2 flex items-center gap-3 text-sm text-app-muted">
            <span>{formatRecord(overview.wins, overview.losses)}</span>
            <span>{formatWinrate(overview.winrate)}</span>
            {stale ? <span className="rounded-full border border-app-warning/40 px-2 py-1 text-app-warning">stale cache</span> : null}
          </div>
        </div>
      </Panel>

      <div className="grid gap-[var(--dashboard-panel-gap)] xl:grid-cols-3">
        <Panel className="min-h-[var(--dashboard-summary-card-min-height)]">
          <div className="mb-4 text-sm font-semibold text-app-muted">Общая статистика</div>
          <div className="flex flex-col items-center gap-4 xl:flex-row xl:items-center xl:justify-between">
            <WinrateRing value={overview.winrate} label="Winrate" />
            <div className="space-y-3">
              <div className="text-sm text-app-muted">Победы</div>
              <div className="text-3xl font-bold text-app-success">{overview.wins}</div>
              <div className="text-sm text-app-muted">Поражения</div>
              <div className="text-2xl font-semibold text-app-danger">{overview.losses}</div>
            </div>
          </div>
        </Panel>

        <Panel className="min-h-[var(--dashboard-summary-card-min-height)]">
          <div className="mb-4 text-sm font-semibold text-app-muted">Матчи</div>
          <div className="space-y-5">
            <div>
              <div className="text-xs uppercase tracking-[0.16em] text-app-muted">Всего матчей</div>
              <div className="mt-2 text-5xl font-bold text-app-text">{overview.totalMatches}</div>
            </div>
            <div className="grid gap-3 rounded-lg border border-app-border/50 bg-app-panelStrong/60 p-4">
              <div className="text-sm text-app-muted">Класс</div>
              <div className="text-xl font-semibold text-app-text">{overview.className}</div>
            </div>
          </div>
        </Panel>

        <Panel className="min-h-[var(--dashboard-summary-card-min-height)]">
          <div className="mb-4 text-sm font-semibold text-app-muted">Лучшие матчапы</div>
          <div className="space-y-3">
            {overview.bestMatchups.map((matchup) => (
              <div key={`${matchup.opponentClass}-${matchup.opponentArchetype}`} className="flex items-center justify-between rounded-lg border border-app-border/50 bg-app-panelStrong/60 px-4 py-3">
                <div>
                  <div className="text-sm font-medium text-app-text">{matchup.opponentArchetype || `Против ${matchup.opponentClass}`}</div>
                  <div className="text-xs text-app-muted">{matchup.matches} матчей</div>
                </div>
                <div className="text-lg font-semibold text-app-success">{formatWinrate(matchup.winrate)}</div>
              </div>
            ))}
          </div>
        </Panel>
      </div>
    </div>
  );
}
