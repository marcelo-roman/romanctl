import { useProjectLogs } from "../../api/queries";
import type { Project } from "../../api/types";
import { StatusDot } from "../../components/Badge";
import { Dialog } from "../../components/Dialog";
import { LogViewer } from "../../components/LogViewer";
import { QueryState } from "../../components/QueryState";
import { useI18n } from "../../i18n/context";

export function ProjectLogsDialog({ project, onClose }: { project: Project; onClose: () => void }) {
  const { t } = useI18n();
  const logs = useProjectLogs(project.name);
  const title = t("projects.logsTitle", { name: project.name });
  return (
    <Dialog title={title} onClose={onClose} size="lg">
      <QueryState query={logs}>
        {(data) => (
          <div className="stack">
            <div className="row-between small">
              <StatusDot on={Boolean(data.running)} label={data.running ? t("projects.logsLive") : t("projects.logsStopped")} />
              {data.path && <code className="mono muted truncate">{data.path}</code>}
            </div>
            <LogViewer lines={data.lines} emptyText={t("projects.logsEmpty")} label={title} />
          </div>
        )}
      </QueryState>
    </Dialog>
  );
}
