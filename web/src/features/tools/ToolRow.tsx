import { useState } from "react";
import { ArrowUpCircle, Download, ListOrdered, Loader2, Trash2 } from "lucide-react";
import type { Tool } from "../../api/types";
import { Badge } from "../../components/Badge";
import { ConfirmDialog } from "../../components/ConfirmDialog";
import { useI18n } from "../../i18n/context";
import { toolJob, useRunningJob } from "../../jobs/useRunningJob";
import { useStartJob } from "../../jobs/useStartJob";
import { hasUpdate } from "../../lib/tools";
import { shortVersion } from "../../lib/version";

function ToolBadges({ tool }: { tool: Tool }) {
  const { t } = useI18n();
  if (tool.unavailable) return <Badge tone="warn">{tool.unavailable}</Badge>;
  return (
    <span className="tags">
      {tool.installed ? <Badge tone="ok">{t("status.installed")}</Badge> : <Badge>{t("status.notInstalled")}</Badge>}
      {tool.installed && tool.version && (
        <Badge title={tool.version}>
          <span className="mono">{shortVersion(tool.version)}</span>
        </Badge>
      )}
      {tool.latest && <span className="muted small">{t("tools.latest", { version: tool.latest })}</span>}
      {hasUpdate(tool) && <Badge tone="info">{t("tools.updateAvailable")}</Badge>}
    </span>
  );
}

function ToolActions({ tool, onChooseVersion }: { tool: Tool; onChooseVersion: () => void }) {
  const { t } = useI18n();
  const start = useStartJob();
  const running = useRunningJob(toolJob(tool.key));
  const [confirming, setConfirming] = useState(false);
  const busy = start.isPending || Boolean(running);
  if (tool.unavailable) return null;
  return (
    <div className="row-actions">
      {busy && <Loader2 size={14} className="spin" aria-label={t("job.status.running")} />}
      {!tool.installed && (
        <button type="button" className="button button-sm" disabled={busy} onClick={() => start.mutate({ action: "tool.install", key: tool.key })}>
          <Download size={13} aria-hidden="true" />
          {t("tools.install")}
        </button>
      )}
      {hasUpdate(tool) && (
        <button type="button" className="button button-sm" disabled={busy} onClick={() => start.mutate({ action: "tool.install", key: tool.key })}>
          <ArrowUpCircle size={13} aria-hidden="true" />
          {t("tools.update")}
        </button>
      )}
      <button type="button" className="button button-sm" disabled={busy} onClick={onChooseVersion}>
        <ListOrdered size={13} aria-hidden="true" />
        {t("tools.chooseVersion")}
      </button>
      {tool.installed && (
        <button type="button" className="button button-sm button-ghost-danger" disabled={busy} onClick={() => setConfirming(true)}>
          <Trash2 size={13} aria-hidden="true" />
          {t("tools.uninstall")}
        </button>
      )}
      {confirming && (
        <ConfirmDialog
          title={t("tools.uninstallTitle", { name: tool.name })}
          confirmLabel={t("tools.uninstall")}
          danger
          onClose={() => setConfirming(false)}
          onConfirm={() => {
            setConfirming(false);
            start.mutate({ action: "tool.uninstall", key: tool.key });
          }}
        >
          <p>{t("tools.uninstallBody", { name: tool.name })}</p>
        </ConfirmDialog>
      )}
    </div>
  );
}

export function ToolRow({ tool, onChooseVersion }: { tool: Tool; onChooseVersion: () => void }) {
  return (
    <li className={tool.unavailable ? "tool-row tool-unavailable" : "tool-row"} aria-disabled={Boolean(tool.unavailable) || undefined}>
      <div className="tool-main">
        <div className="row">
          <span className="strong">{tool.name}</span>
          <code className="mono muted small">{tool.key}</code>
        </div>
        <span className="muted small">{tool.description}</span>
      </div>
      <ToolBadges tool={tool} />
      <ToolActions tool={tool} onChooseVersion={onChooseVersion} />
    </li>
  );
}
