import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useJob } from "../api/queries";
import { useToast } from "../feedback/toastContext";
import { useI18n } from "../i18n/context";
import { isFinished, relatedQueryKeys } from "../lib/jobs";
import { JobTrackerContext, type JobTracker } from "./jobTrackerContext";

function TrackedJob({ id, onDone }: { id: string; onDone: (id: string) => void }) {
  const { data: job } = useJob(id);
  const queryClient = useQueryClient();
  const toast = useToast();
  const { t } = useI18n();
  const handled = useRef(false);

  useEffect(() => {
    if (!job || !isFinished(job) || handled.current) return;
    handled.current = true;
    for (const key of relatedQueryKeys(job.action)) void queryClient.invalidateQueries({ queryKey: key });
    const succeeded = job.status === "succeeded";
    toast.push({
      kind: succeeded ? "success" : "error",
      message: t(succeeded ? "jobs.toastSucceeded" : "jobs.toastFailed", { title: job.title }),
      link: { to: `/jobs/${job.id}`, label: t("jobs.view") },
    });
    onDone(id);
  }, [job, id, onDone, queryClient, toast, t]);

  return null;
}

export function JobTrackerProvider({ children }: { children: ReactNode }) {
  const [tracked, setTracked] = useState<string[]>([]);

  const track = useCallback((id: string) => {
    setTracked((current) => (current.includes(id) ? current : [...current, id]));
  }, []);

  const untrack = useCallback((id: string) => {
    setTracked((current) => current.filter((item) => item !== id));
  }, []);

  const value = useMemo<JobTracker>(() => ({ tracked, track, untrack }), [tracked, track, untrack]);

  return (
    <JobTrackerContext.Provider value={value}>
      {children}
      {tracked.map((id) => (
        <TrackedJob key={id} id={id} onDone={untrack} />
      ))}
    </JobTrackerContext.Provider>
  );
}
