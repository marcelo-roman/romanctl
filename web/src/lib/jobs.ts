import type { QueryKey } from "@tanstack/react-query";
import { queryKeys } from "../api/queries";
import type { Job, JobRequest, Template } from "../api/types";

export type ProjectSource = { kind: "simple"; type: string } | { kind: "template"; template: Template };

export interface NewProjectInput {
  source: ProjectSource;
  name: string;
  services: string[];
  git: boolean;
}

export function buildCreateBody(input: NewProjectInput): JobRequest {
  if (input.source.kind === "simple") {
    return { action: "project.create", name: input.name, type: input.source.type, git: input.git };
  }
  const { template } = input.source;
  return {
    action: "project.create",
    name: input.name,
    stack: template.stack,
    arch: [...template.arch],
    with: [...input.services],
    git: input.git,
  };
}

export function jobCommand(job: Pick<Job, "args">): string {
  return ["romanctl", ...job.args].join(" ");
}

export function relatedQueryKeys(action: string): QueryKey[] {
  const keys: QueryKey[] = [queryKeys.jobs];
  if (action.startsWith("project.")) keys.push(queryKeys.projects);
  if (action.startsWith("project.up") || action.startsWith("project.down")) keys.push(queryKeys.services);
  if (action.startsWith("service.")) keys.push(queryKeys.services, queryKeys.projects);
  if (action.startsWith("tool.")) keys.push(queryKeys.tools);
  if (action.startsWith("templates.")) keys.push(queryKeys.info, queryKeys.catalog, queryKeys.projects);
  return keys;
}

export function isFinished(job: Pick<Job, "status">): boolean {
  return job.status !== "running";
}
