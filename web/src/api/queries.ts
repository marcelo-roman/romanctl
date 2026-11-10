import { useQuery } from "@tanstack/react-query";
import { api } from "./client";
import type { Job } from "./types";

export const queryKeys = {
  info: ["info"] as const,
  system: ["system"] as const,
  projects: ["projects"] as const,
  projectLogs: (name: string) => ["projects", name, "logs"] as const,
  catalog: ["catalog"] as const,
  services: ["services"] as const,
  serviceLogs: (key: string) => ["services", key, "logs"] as const,
  tools: ["tools"] as const,
  toolVersions: (key: string) => ["tools", key, "versions"] as const,
  jobs: ["jobs"] as const,
  job: (id: string) => ["jobs", id] as const,
};

export const systemRefreshMs = 5000;
export const jobPollMs = 1000;
export const logPollMs = 2000;

export function useInfo() {
  return useQuery({ queryKey: queryKeys.info, queryFn: api.info, staleTime: 60_000 });
}

export function useSystem() {
  return useQuery({ queryKey: queryKeys.system, queryFn: api.system, refetchInterval: systemRefreshMs });
}

export function useProjects() {
  return useQuery({ queryKey: queryKeys.projects, queryFn: api.projects, refetchInterval: 10_000 });
}

export function useProjectLogs(name: string) {
  return useQuery({
    queryKey: queryKeys.projectLogs(name),
    queryFn: () => api.projectLogs(name),
    refetchInterval: (query) => (query.state.data?.running ? logPollMs : false),
  });
}

export function useCatalog() {
  return useQuery({ queryKey: queryKeys.catalog, queryFn: api.catalog, staleTime: 60_000 });
}

export function useServices() {
  return useQuery({ queryKey: queryKeys.services, queryFn: api.services, refetchInterval: 15_000 });
}

export function useServiceLogs(key: string, live: boolean) {
  return useQuery({
    queryKey: queryKeys.serviceLogs(key),
    queryFn: () => api.serviceLogs(key),
    refetchInterval: live ? logPollMs * 2 : false,
  });
}

export function useTools() {
  return useQuery({ queryKey: queryKeys.tools, queryFn: api.tools, staleTime: 60_000 });
}

export function useToolVersions(key: string) {
  return useQuery({ queryKey: queryKeys.toolVersions(key), queryFn: () => api.toolVersions(key), staleTime: 300_000 });
}

function anyRunning(jobs: Job[] | undefined): boolean {
  return Boolean(jobs?.some((job) => job.status === "running"));
}

export function useJobs() {
  return useQuery({
    queryKey: queryKeys.jobs,
    queryFn: api.jobs,
    refetchInterval: (query) => (anyRunning(query.state.data) ? jobPollMs * 2 : 10_000),
  });
}

export function useJob(id: string) {
  return useQuery({
    queryKey: queryKeys.job(id),
    queryFn: () => api.job(id),
    refetchInterval: (query) => (query.state.data?.status === "running" ? jobPollMs : false),
  });
}
