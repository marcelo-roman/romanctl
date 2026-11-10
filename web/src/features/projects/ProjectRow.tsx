import { Code2, Loader2, Play, ScrollText, Square, Trash2 } from "lucide-react";
import type { Project } from "../../api/types";
import { Badge, StatusDot } from "../../components/Badge";
import { ExternalLink } from "../../components/ExternalLink";
import { useI18n } from "../../i18n/context";
import { useStartJob } from "../../jobs/useStartJob";
import { projectJob, useRunningJob } from "../../jobs/useRunningJob";
import { isProjectUp, projectKind, projectUrl } from "../../lib/catalog";
import type { ProjectDialog } from "./ProjectsPage";

interface ProjectRowProps {
  project: Project;
  onOpen: (dialog: ProjectDialog) => void;
}

function ArchTags({ project }: { project: Project }) {
  const arch = project.manifest?.arch ?? [];
  if (arch.length === 0) return null;
  return (
    <span className="tags">
      {arch.map((tag) => (
        <Badge key={tag}>{tag}</Badge>
      ))}
    </span>
  );
}

export function ProjectRow({ project, onOpen }: ProjectRowProps) {
  const { t } = useI18n();
  const start = useStartJob();
  const up = isProjectUp(project);
  const name = project.name;
  const running = useRunningJob(projectJob(name));
  const busy = start.isPending || Boolean(running);

  return (
    <tr>
      <th scope="row" className="strong">
        {name}
      </th>
      <td>
        <div className="cell-stack">
          <code className="mono small">{projectKind(project)}</code>
          <ArchTags project={project} />
        </div>
      </td>
      <td>
        {busy ? <Loader2 size={14} className="spin" aria-label={t("job.status.running")} /> : <StatusDot on={up} label={up ? t("status.up") : t("status.down")} />}
      </td>
      <td>
        <ExternalLink url={projectUrl(project)} />
      </td>
      <td>
        <div className="row-actions" role="group" aria-label={t("projects.actionsFor", { name })}>
          {up ? (
            <button type="button" className="button button-sm" disabled={busy} onClick={() => start.mutate({ action: "project.down", name })}>
              <Square size={13} aria-hidden="true" />
              {t("projects.down")}
            </button>
          ) : (
            <button type="button" className="button button-sm" disabled={busy} onClick={() => start.mutate({ action: "project.up", name })}>
              <Play size={13} aria-hidden="true" />
              {t("projects.up")}
            </button>
          )}
          <button type="button" className="button button-sm" onClick={() => onOpen({ kind: "logs", project })}>
            <ScrollText size={13} aria-hidden="true" />
            {t("common.logs")}
          </button>
          <button
            type="button"
            className="button button-sm"
            disabled={project.generators.length === 0}
            onClick={() => onOpen({ kind: "generate", project })}
          >
            <Code2 size={13} aria-hidden="true" />
            {t("projects.generate")}
          </button>
          <button type="button" className="button button-sm button-ghost-danger" onClick={() => onOpen({ kind: "remove", project })}>
            <Trash2 size={13} aria-hidden="true" />
            {t("projects.remove")}
          </button>
        </div>
      </td>
    </tr>
  );
}
