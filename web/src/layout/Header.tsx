import { Link } from "react-router";
import { Loader2, Menu } from "lucide-react";
import { useInfo, useJobs } from "../api/queries";
import type { Lang, PlatformInfo } from "../api/types";
import { useI18n } from "../i18n/context";

function platformLabel(platform: PlatformInfo | undefined): string {
  if (!platform) return "";
  const extras = [platform.wsl ? "WSL" : "", platform.proot ? "proot" : ""].filter(Boolean);
  return [`${platform.os}/${platform.arch}`, ...extras].join(" · ");
}

function RunningJobs() {
  const { t } = useI18n();
  const jobs = useJobs();
  const running = jobs.data?.filter((job) => job.status === "running").length ?? 0;
  if (running === 0) return null;
  return (
    <Link to="/jobs" className="running-indicator">
      <Loader2 size={14} className="spin" aria-hidden="true" />
      {t("header.runningJobs", { count: running })}
    </Link>
  );
}

function LanguageSwitch() {
  const { lang, setLang, t } = useI18n();
  return (
    <label className="lang-switch">
      <span className="sr-only">{t("header.language")}</span>
      <select className="select select-sm" value={lang} onChange={(event) => setLang(event.target.value as Lang)}>
        <option value="pt">PT</option>
        <option value="en">EN</option>
      </select>
    </label>
  );
}

export function Header({ menuOpen, onToggleMenu }: { menuOpen: boolean; onToggleMenu: () => void }) {
  const { t } = useI18n();
  const info = useInfo();
  return (
    <header className="topbar">
      <button
        type="button"
        className="icon-button menu-button"
        aria-label={menuOpen ? t("nav.closeMenu") : t("nav.openMenu")}
        aria-expanded={menuOpen}
        aria-controls="sidebar"
        onClick={onToggleMenu}
      >
        <Menu size={18} aria-hidden="true" />
      </button>
      <div className="topbar-meta">
        {info.data && <span className="badge badge-neutral mono">v{info.data.version}</span>}
        <span className="muted mono">{platformLabel(info.data?.platform)}</span>
      </div>
      <div className="topbar-actions">
        <RunningJobs />
        <LanguageSwitch />
      </div>
    </header>
  );
}
