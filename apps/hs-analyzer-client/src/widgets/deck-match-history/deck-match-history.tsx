import { useState } from "react";
import type { DeckHistoryEntry } from "@/shared/model/dashboard";
import { Panel } from "@/shared/ui/panel";
import { Pill } from "@/shared/ui/pill";

function formatPlayedAt(value: string): string {
  if (!value) {
    return "Время неизвестно";
  }

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(parsed);
}

function historyTone(result: string): "success" | "danger" | "warning" {
  switch (result) {
    case "win":
      return "success";
    case "loss":
      return "danger";
    default:
      return "warning";
  }
}

function resultLabel(result: string): string {
  switch (result) {
    case "win":
      return "Победа";
    case "loss":
      return "Поражение";
    default:
      return "Неизвестно";
  }
}

function opponentLabel(opponentClass: string): string {
  return opponentClass || "UNKNOWN";
}

export function DeckMatchHistory({ history }: { history: DeckHistoryEntry[] }) {
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
          <div className="text-xs uppercase tracking-[0.2em] text-app-muted">Timeline</div>
          <div className="text-xl font-semibold text-app-text">История матчей</div>
        </div>
        <div className="flex items-center gap-3">
          <Pill>{history.length}</Pill>
          <span className="text-xs uppercase tracking-[0.18em] text-app-muted">{open ? "Hide" : "Show"}</span>
        </div>
      </button>
      {!open ? null : history.length === 0 ? <div className="text-sm text-app-muted">История для этой колоды пока пустая.</div> : null}
      {!open ? null : <div className="space-y-3">
        {history.map((entry) => (
          <div key={entry.matchId} className="grid gap-3 rounded-lg border border-app-border/50 bg-app-panelStrong/60 p-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
            <div className="space-y-1">
              <div className="text-sm font-medium text-app-text">Против {opponentLabel(entry.opponentClass)}</div>
              <div className="text-xs text-app-muted">{formatPlayedAt(entry.playedAt)}</div>
            </div>
            <Pill tone={historyTone(entry.result)}>{resultLabel(entry.result)}</Pill>
          </div>
        ))}
      </div>}
    </Panel>
  );
}
