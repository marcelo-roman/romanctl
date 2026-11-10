import { useState } from "react";
import { Search } from "lucide-react";
import { useTools } from "../../api/queries";
import type { Tool, ToolList } from "../../api/types";
import { PageHeader } from "../../components/PageHeader";
import { QueryState } from "../../components/QueryState";
import { useI18n } from "../../i18n/context";
import { allGroups, countByGroup } from "../../lib/tools";
import { ToolRow } from "./ToolRow";
import { ToolVersionsDialog } from "./ToolVersionsDialog";
import { useToolFilter } from "./useToolFilter";

type Filter = ReturnType<typeof useToolFilter>;

function GroupChips({ data, filter }: { data: ToolList; filter: Filter }) {
  const { t } = useI18n();
  const counts = countByGroup(data.tools);
  const chips = [{ key: allGroups, title: t("common.all"), count: data.tools.length }, ...data.groups.map((group) => ({ ...group, count: counts[group.key] ?? 0 }))];
  return (
    <div className="chips" role="group" aria-label={t("tools.groups")}>
      {chips.map((chip) => (
        <button
          key={chip.key}
          type="button"
          className={filter.group === chip.key ? "chip chip-active" : "chip"}
          aria-pressed={filter.group === chip.key}
          onClick={() => filter.setGroup(chip.key)}
        >
          {chip.title}
          <span className="chip-count">{chip.count}</span>
        </button>
      ))}
    </div>
  );
}

function Toolbar({ data, filter }: { data: ToolList; filter: Filter }) {
  const { t } = useI18n();
  return (
    <div className="toolbar">
      <label className="search">
        <Search size={14} aria-hidden="true" />
        <span className="sr-only">{t("tools.search")}</span>
        <input
          type="search"
          className="input"
          placeholder={t("tools.searchPlaceholder")}
          value={filter.search}
          onChange={(event) => filter.setSearch(event.target.value)}
        />
      </label>
      <label className="checkbox">
        <input type="checkbox" checked={filter.installedOnly} onChange={(event) => filter.setInstalledOnly(event.target.checked)} />
        {t("tools.installedOnly")}
      </label>
      <span className="muted small" aria-live="polite">
        {t("tools.count", { shown: filter.visible.length, total: data.tools.length })}
      </span>
    </div>
  );
}

function ToolCatalog({ data }: { data: ToolList }) {
  const { t } = useI18n();
  const filter = useToolFilter(data.groups, data.tools);
  const [versionsFor, setVersionsFor] = useState<Tool | null>(null);
  return (
    <div className="stack">
      <Toolbar data={data} filter={filter} />
      <GroupChips data={data} filter={filter} />
      {filter.sections.length === 0 && <p className="muted">{t("tools.empty")}</p>}
      {filter.sections.map((section) => (
        <section key={section.group.key} className="panel panel-flush" aria-labelledby={`group-${section.group.key}`}>
          <h2 id={`group-${section.group.key}`} className="panel-title">
            {section.group.title}
          </h2>
          <ul className="tool-list">
            {section.tools.map((tool) => (
              <ToolRow key={tool.key} tool={tool} onChooseVersion={() => setVersionsFor(tool)} />
            ))}
          </ul>
        </section>
      ))}
      {versionsFor && <ToolVersionsDialog tool={versionsFor} onClose={() => setVersionsFor(null)} />}
    </div>
  );
}

export function ToolsPage() {
  const { t } = useI18n();
  const tools = useTools();
  return (
    <>
      <PageHeader title={t("tools.title")} />
      <QueryState query={tools} loadingText={t("tools.loadingSlow")}>
        {(data) => <ToolCatalog data={data} />}
      </QueryState>
    </>
  );
}
