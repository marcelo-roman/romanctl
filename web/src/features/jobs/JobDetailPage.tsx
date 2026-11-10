import { Link, useParams } from "react-router";
import { ArrowLeft } from "lucide-react";
import { useJob } from "../../api/queries";
import type { Job } from "../../api/types";
import { CopyButton } from "../../components/CopyButton";
import { LogViewer } from "../../components/LogViewer";
import { PageHeader, Section } from "../../components/PageHeader";
import { QueryState } from "../../components/QueryState";
import { useI18n } from "../../i18n/context";
import { formatDuration, formatTime, prettyJson } from "../../lib/format";
import { jobCommand } from "../../lib/jobs";
import { JobStatusBadge } from "./JobStatusBadge";

function JobMeta({ job }: { job: Job }) {
  const { t, lang } = useI18n();
  const command = jobCommand(job);
  return (
    <dl className="meta-grid">
      <dt>{t("projects.status")}</dt>
      <dd>
        <JobStatusBadge status={job.status} />
      </dd>
      <dt>{t("jobs.started")}</dt>
      <dd className="mono">{formatTime(job.started, lang)}</dd>
      <dt>{t("jobs.duration")}</dt>
      <dd className="mono">{formatDuration(job.started, job.finished)}</dd>
      <dt>{t("jobs.command")}</dt>
      <dd className="row">
        <code className="command">{command}</code>
        <CopyButton value={command} />
      </dd>
    </dl>
  );
}

function JobResult({ result }: { result: unknown }) {
  const { t } = useI18n();
  return (
    <details className="panel result" open>
      <summary>{t("jobs.result")}</summary>
      <pre className="log json">{prettyJson(result)}</pre>
    </details>
  );
}

function JobView({ job }: { job: Job }) {
  const { t } = useI18n();
  return (
    <div className="stack">
      <section className="panel">
        <JobMeta job={job} />
      </section>
      {job.error && (
        <div className="error-box" role="alert">
          <strong>{t("jobs.error")}:</strong> {job.error}
        </div>
      )}
      <Section title={t("jobs.output")}>
        <LogViewer lines={job.lines} emptyText={t("jobs.noOutput")} label={t("jobs.output")} />
      </Section>
      {job.result !== undefined && <JobResult result={job.result} />}
    </div>
  );
}

export function JobDetailPage() {
  const { t } = useI18n();
  const { id = "" } = useParams();
  const job = useJob(id);
  return (
    <>
      <Link className="link small back-link" to="/jobs">
        <ArrowLeft size={14} aria-hidden="true" />
        {t("jobs.title")}
      </Link>
      <PageHeader title={job.data?.title ?? t("jobs.task")} />
      <QueryState query={job}>{(data) => <JobView job={data} />}</QueryState>
    </>
  );
}
