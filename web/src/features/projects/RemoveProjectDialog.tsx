import type { Project } from "../../api/types";
import { ConfirmDialog } from "../../components/ConfirmDialog";
import { useI18n } from "../../i18n/context";
import { useStartJob } from "../../jobs/useStartJob";

export function RemoveProjectDialog({ project, onClose }: { project: Project; onClose: () => void }) {
  const { t } = useI18n();
  const start = useStartJob({ onStarted: onClose });
  return (
    <ConfirmDialog
      title={t("projects.removeTitle", { name: project.name })}
      confirmLabel={t("projects.remove")}
      danger
      typeToConfirm={project.name}
      busy={start.isPending}
      onClose={onClose}
      onConfirm={() => start.mutate({ action: "project.remove", name: project.name })}
    >
      <p>{t("projects.removeWarning", { path: project.path })}</p>
    </ConfirmDialog>
  );
}
