import { useState } from "react";
import { Link } from "react-router";
import { Plus } from "lucide-react";
import { useInfo, useProjects } from "../../api/queries";
import type { Project } from "../../api/types";
import { PageHeader } from "../../components/PageHeader";
import { QueryState } from "../../components/QueryState";
import { useI18n } from "../../i18n/context";
import { GenerateDialog } from "./GenerateDialog";
import { ProjectLogsDialog } from "./ProjectLogsDialog";
import { ProjectRow } from "./ProjectRow";
import { RemoveProjectDialog } from "./RemoveProjectDialog";

export type ProjectDialog = { kind: "logs" | "generate" | "remove"; project: Project } | null;

function ActiveDialog({ dialog, onClose }: { dialog: ProjectDialog; onClose: () => void }) {
  if (!dialog) return null;
  if (dialog.kind === "logs") return <ProjectLogsDialog project={dialog.project} onClose={onClose} />;
  if (dialog.kind === "generate") return <GenerateDialog project={dialog.project} onClose={onClose} />;
  return <RemoveProjectDialog project={dialog.project} onClose={onClose} />;
}

function ProjectTable({ projects, onOpen }: { projects: Project[]; onOpen: (dialog: ProjectDialog) => void }) {
  const { t } = useI18n();
  const info = useInfo();
  if (projects.length === 0) return <p className="muted">{t("projects.empty", { dir: info.data?.projectsDir ?? "" })}</p>;
  return (
    <div className="table-wrap">
      <table className="table">
        <thead>
          <tr>
            <th scope="col">{t("projects.name")}</th>
            <th scope="col">{t("projects.kind")}</th>
            <th scope="col">{t("projects.status")}</th>
            <th scope="col">{t("projects.url")}</th>
            <th scope="col" className="actions-col">
              {t("projects.actions")}
            </th>
          </tr>
        </thead>
        <tbody>
          {projects.map((project) => (
            <ProjectRow key={project.name} project={project} onOpen={onOpen} />
          ))}
        </tbody>
      </table>
    </div>
  );
}

export function ProjectsPage() {
  const { t } = useI18n();
  const projects = useProjects();
  const [dialog, setDialog] = useState<ProjectDialog>(null);
  return (
    <>
      <PageHeader title={t("projects.title")}>
        <Link className="button button-primary" to="/projects/new">
          <Plus size={14} aria-hidden="true" />
          {t("projects.new")}
        </Link>
      </PageHeader>
      <section className="panel panel-flush">
        <QueryState query={projects}>{(list) => <ProjectTable projects={list} onOpen={setDialog} />}</QueryState>
      </section>
      <ActiveDialog dialog={dialog} onClose={() => setDialog(null)} />
    </>
  );
}
