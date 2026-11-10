import { useToolVersions } from "../../api/queries";
import type { Tool } from "../../api/types";
import { Dialog } from "../../components/Dialog";
import { QueryState } from "../../components/QueryState";
import { useI18n } from "../../i18n/context";
import { useStartJob } from "../../jobs/useStartJob";

export function ToolVersionsDialog({ tool, onClose }: { tool: Tool; onClose: () => void }) {
  const { t } = useI18n();
  const versions = useToolVersions(tool.key);
  const start = useStartJob({ onStarted: onClose });
  return (
    <Dialog title={t("tools.versionsTitle", { name: tool.name })} onClose={onClose}>
      <QueryState query={versions}>
        {(data) => (
          <div className="stack">
            {data.error && (
              <p className="error-box" role="alert">
                {data.error}
              </p>
            )}
            {data.versions.length === 0 && !data.error && <p className="muted">{t("tools.versionsEmpty")}</p>}
            <ul className="version-list">
              {data.versions.map((version) => (
                <li key={version.id} className="row-between">
                  <span>
                    <span className="mono strong">{version.label}</span>
                    {version.note && <span className="muted small"> · {version.note}</span>}
                  </span>
                  <button
                    type="button"
                    className="button button-sm"
                    disabled={start.isPending}
                    aria-label={t("tools.installVersion", { version: version.label })}
                    onClick={() => start.mutate({ action: "tool.install", key: tool.key, version: version.id })}
                  >
                    {t("tools.install")}
                  </button>
                </li>
              ))}
            </ul>
          </div>
        )}
      </QueryState>
    </Dialog>
  );
}
