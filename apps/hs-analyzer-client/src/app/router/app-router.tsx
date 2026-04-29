import { HashRouter, NavLink, Route, Routes } from "react-router-dom";
import { DeckOverviewPage } from "@/pages/deck-overview/page";
import { LiveOverlayControlPage } from "@/pages/live-overlay-control/page";
import { SettingsPage } from "@/pages/settings/page";

export function AppRouter() {
  return (
    <HashRouter>
      <div className="border-b border-app-border/50 bg-black/20 px-6 py-3">
        <div className="flex gap-4 text-sm text-app-muted">
          <NavLink to="/" end className={({ isActive }) => (isActive ? "text-app-text" : "")}>
            Dashboard
          </NavLink>
          <NavLink to="/overlay" className={({ isActive }) => (isActive ? "text-app-text" : "")}>
            Overlay
          </NavLink>
          <NavLink to="/settings" className={({ isActive }) => (isActive ? "text-app-text" : "")}>
            Settings
          </NavLink>
        </div>
      </div>
      <Routes>
        <Route path="/" element={<DeckOverviewPage />} />
        <Route path="/overlay" element={<LiveOverlayControlPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Routes>
    </HashRouter>
  );
}
