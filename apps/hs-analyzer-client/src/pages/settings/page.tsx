import { useEffect, useState } from "react";
import type { OverlayMode } from "@electron/shared/contracts";
import { useServices } from "@/shared/api/services-context";
import { Button } from "@/shared/ui/button";
import { Pill } from "@/shared/ui/pill";

export function SettingsPage() {
  const { bridge } = useServices();
  const [mode, setMode] = useState<OverlayMode>("inspect");
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    let cancelled = false;

    void bridge.getOverlayMode().then((nextMode) => {
      if (!cancelled) {
        setMode(nextMode);
      }
    });

    return () => {
      cancelled = true;
    };
  }, [bridge]);

  const inspectEnabled = mode === "inspect";

  const toggleInspectMode = async () => {
    const nextMode: OverlayMode = inspectEnabled ? "passive" : "inspect";
    setIsSaving(true);
    try {
      const resolvedMode = await bridge.setOverlayMode(nextMode);
      setMode(resolvedMode);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="min-h-screen p-6">
      <div className="space-y-6 rounded-xl border border-app-border/70 bg-app-panel/60 p-6">
        <div className="text-xs uppercase tracking-[0.2em] text-app-muted">Settings</div>
        <div className="mt-3 text-2xl font-bold text-app-text">Настройки клиента</div>
        <div className="mt-2 text-sm text-app-muted">Управление поведением игрового оверлея и live diagnostics.</div>

        <section className="rounded-lg border border-app-border/70 bg-app-panelStrong/55 p-4">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="max-w-2xl">
              <div className="flex flex-wrap items-center gap-3">
                <h2 className="text-lg font-semibold text-app-text">Inspect mode поверх Hearthstone</h2>
                <Pill tone={inspectEnabled ? "success" : "warning"}>{inspectEnabled ? "включен" : "выключен"}</Pill>
              </div>
              <p className="mt-2 text-sm leading-6 text-app-muted">
                По умолчанию inspect mode включен, чтобы hover по картам открывал полное описание прямо в игровом оверлее.
                Если отключить его, окно снова станет click-through и не будет перехватывать мышь.
              </p>
            </div>
            <Button
              type="button"
              role="switch"
              aria-checked={inspectEnabled}
              disabled={isSaving}
              onClick={() => void toggleInspectMode()}
            >
              {inspectEnabled ? "Отключить inspect mode" : "Включить inspect mode"}
            </Button>
          </div>
        </section>
      </div>
    </div>
  );
}
