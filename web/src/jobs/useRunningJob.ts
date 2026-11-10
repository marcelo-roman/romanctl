import { useJobs } from "../api/queries";
import type { Job } from "../api/types";

export type JobMatcher = (job: Job) => boolean;

export function projectJob(name: string): JobMatcher {
  return (job) => ["project.up", "project.down", "project.remove"].includes(job.action) && job.args[1] === name;
}

export function serviceJob(key: string): JobMatcher {
  return (job) => job.action.startsWith("service.") && job.args[1] === key;
}

export function toolJob(key: string): JobMatcher {
  return (job) => job.action.startsWith("tool.") && job.args[2] === key;
}

export function useRunningJob(match: JobMatcher): Job | undefined {
  const jobs = useJobs();
  return jobs.data?.find((job) => job.status === "running" && match(job));
}
