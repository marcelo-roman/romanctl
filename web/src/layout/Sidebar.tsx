import { NavLink } from "react-router";
import { Boxes, FolderGit2, LayoutDashboard, LayoutTemplate, ListChecks, Server, Wrench } from "lucide-react";
import type { MessageKey } from "../i18n/messages";
import { useI18n } from "../i18n/context";

interface NavItem {
  to: string;
  label: MessageKey;
  icon: typeof Boxes;
  end?: boolean;
}

const items: NavItem[] = [
  { to: "/", label: "nav.dashboard", icon: LayoutDashboard, end: true },
  { to: "/projects", label: "nav.projects", icon: FolderGit2 },
  { to: "/services", label: "nav.services", icon: Server },
  { to: "/tools", label: "nav.tools", icon: Wrench },
  { to: "/templates", label: "nav.templates", icon: LayoutTemplate },
  { to: "/jobs", label: "nav.jobs", icon: ListChecks },
];

export function Sidebar({ open, onNavigate }: { open: boolean; onNavigate: () => void }) {
  const { t } = useI18n();
  return (
    <nav id="sidebar" className={open ? "sidebar sidebar-open" : "sidebar"} aria-label={t("nav.main")}>
      <div className="brand">
        <Boxes size={20} aria-hidden="true" />
        <span>{t("app.name")}</span>
      </div>
      <ul>
        {items.map(({ to, label, icon: Icon, end }) => (
          <li key={to}>
            <NavLink to={to} end={end} className="nav-link" onClick={onNavigate}>
              <Icon size={16} aria-hidden="true" />
              <span>{t(label)}</span>
            </NavLink>
          </li>
        ))}
      </ul>
    </nav>
  );
}
