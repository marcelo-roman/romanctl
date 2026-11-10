import { useState } from "react";
import { Outlet } from "react-router";
import { useI18n } from "../i18n/context";
import { Header } from "./Header";
import { Sidebar } from "./Sidebar";

export function Layout() {
  const { t } = useI18n();
  const [menuOpen, setMenuOpen] = useState(false);
  return (
    <div className="shell">
      <a className="skip-link" href="#main">
        {t("app.skipToContent")}
      </a>
      <Sidebar open={menuOpen} onNavigate={() => setMenuOpen(false)} />
      <div className="main-column">
        <Header menuOpen={menuOpen} onToggleMenu={() => setMenuOpen((open) => !open)} />
        <main id="main" className="content" tabIndex={-1}>
          <Outlet />
        </main>
      </div>
    </div>
  );
}
