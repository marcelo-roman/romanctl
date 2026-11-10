import { Link } from "react-router";
import type { Job } from "../../api/types";
import { useI18n } from "../../i18n/context";
import { formatDuration, formatTime } from "../../lib/format";
import { JobStatusBadge } from "./JobStatusBadge";

export function JobList({ jobs, emptyText }: { jobs: Job[]; emptyText: string }) {
  const { lang } = useI18n();
  if (jobs.length === 0) return <p className="muted">{emptyText}</p>;
  return (
    <ul className="job-list">
      {jobs.map((job) => (
        <li key={job.id}>
          <Link className="job-item" to={`/jobs/${job.id}`}>
            <JobStatusBadge status={job.status} />
            <span className="job-title">{job.title}</span>
            <span className="muted small mono">{formatTime(job.started, lang)}</span>
            <span className="muted small mono">{formatDuration(job.started, job.finished)}</span>
          </Link>
        </li>
      ))}
    </ul>
  );
}
