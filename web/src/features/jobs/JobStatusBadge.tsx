import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import type { JobStatus } from "../../api/types";
import { Badge, type Tone } from "../../components/Badge";
import { useI18n } from "../../i18n/context";

const tones: Record<JobStatus, Tone> = { running: "info", succeeded: "ok", failed: "danger" };
const icons = { running: Loader2, succeeded: CheckCircle2, failed: XCircle };

export function JobStatusBadge({ status }: { status: JobStatus }) {
  const { t } = useI18n();
  const Icon = icons[status];
  return (
    <Badge tone={tones[status]}>
      <Icon size={12} className={status === "running" ? "spin" : undefined} aria-hidden="true" />
      {t(`job.status.${status}`)}
    </Badge>
  );
}
