import { createContext, useContext, useEffect, useMemo, useState, type PropsWithChildren } from "react";
import { DashboardClient } from "@/shared/api/dashboard-client";
import { LiveOverlayClient } from "@/shared/api/live-client";
import { getNativeBridge } from "@/shared/config/app-config";

interface Services {
  dashboardClient: DashboardClient;
  liveClient: LiveOverlayClient;
  bridge: ReturnType<typeof getNativeBridge>;
}

const ServicesContext = createContext<Services | null>(null);

export function ServicesProvider({ children }: PropsWithChildren) {
  const bridge = getNativeBridge();
  const [config, setConfig] = useState({
    backendBaseUrl: "http://127.0.0.1:8080",
    liveBaseUrl: "http://127.0.0.1:8765",
  });

  useEffect(() => {
    let cancelled = false;
    void bridge
      .getConfig()
      .then((nextConfig) => {
        if (!cancelled) {
          setConfig((currentConfig) => {
            if (
              currentConfig.backendBaseUrl === nextConfig.backendBaseUrl &&
              currentConfig.liveBaseUrl === nextConfig.liveBaseUrl
            ) {
              return currentConfig;
            }
            return nextConfig;
          });
        }
      })
      .catch(() => undefined);

    return () => {
      cancelled = true;
    };
  }, [bridge]);

  const services = useMemo<Services>(
    () => ({
      dashboardClient: new DashboardClient(config.backendBaseUrl),
      liveClient: new LiveOverlayClient(config.liveBaseUrl),
      bridge,
    }),
    [bridge, config.backendBaseUrl, config.liveBaseUrl],
  );

  return <ServicesContext.Provider value={services}>{children}</ServicesContext.Provider>;
}

export function useServices() {
  const services = useContext(ServicesContext);
  if (!services) {
    throw new Error("Services context is not available");
  }
  return services;
}
