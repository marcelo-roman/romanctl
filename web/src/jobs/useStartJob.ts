import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router";
import { api } from "../api/client";
import { queryKeys } from "../api/queries";
import type { Job, JobRequest } from "../api/types";
import { useToast } from "../feedback/toastContext";
import { useI18n } from "../i18n/context";
import { useJobTracker } from "./jobTrackerContext";

interface StartOptions {
  navigateToJob?: boolean;
  onStarted?: (job: Job) => void;
}

export function useStartJob(options: StartOptions = {}) {
  const queryClient = useQueryClient();
  const tracker = useJobTracker();
  const toast = useToast();
  const navigate = useNavigate();
  const { t } = useI18n();

  return useMutation({
    mutationFn: (body: JobRequest) => api.startJob(body),
    onSuccess: (job) => {
      queryClient.setQueryData(queryKeys.job(job.id), job);
      void queryClient.invalidateQueries({ queryKey: queryKeys.jobs, exact: true });
      tracker.track(job.id);
      options.onStarted?.(job);
      if (options.navigateToJob) {
        navigate(`/jobs/${job.id}`);
        return;
      }
      toast.push({ kind: "info", message: t("jobs.started.toast", { title: job.title }), link: { to: `/jobs/${job.id}`, label: t("jobs.view") } });
    },
    onError: (error) => {
      toast.push({ kind: "error", message: error.message });
    },
  });
}
