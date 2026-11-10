import type {
  Catalog,
  Info,
  Job,
  JobRequest,
  LogLines,
  Project,
  Service,
  SystemSnapshot,
  ToolList,
  ToolVersions,
} from "./types";

export const tokenHeader = "X-Romanctl-Token";

export class ApiError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export function readToken(): string {
  const meta = document.querySelector<HTMLMetaElement>('meta[name="romanctl-token"]');
  if (meta?.content) return meta.content;
  return import.meta.env.VITE_ROMANCTL_TOKEN ?? "";
}

async function errorMessage(response: Response): Promise<string> {
  const text = await response.text();
  try {
    const body = JSON.parse(text) as { error?: string };
    if (body.error) return body.error;
  } catch {
    if (text.trim()) return text.trim();
  }
  return `HTTP ${response.status}`;
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set(tokenHeader, readToken());
  headers.set("Accept", "application/json");
  if (init.body) headers.set("Content-Type", "application/json");
  const response = await fetch(new URL(path, window.location.origin), { ...init, headers });
  if (!response.ok) throw new ApiError(await errorMessage(response), response.status);
  return (await response.json()) as T;
}

const encode = encodeURIComponent;

export const api = {
  info: () => request<Info>("/api/info"),
  system: () => request<SystemSnapshot>("/api/system"),
  projects: () => request<Project[]>("/api/projects"),
  projectLogs: (name: string, lines = 300) => request<LogLines>(`/api/projects/${encode(name)}/logs?lines=${lines}`),
  catalog: () => request<Catalog>("/api/catalog"),
  services: () => request<Service[]>("/api/services"),
  serviceLogs: (key: string, lines = 300) => request<LogLines>(`/api/services/${encode(key)}/logs?lines=${lines}`),
  tools: () => request<ToolList>("/api/tools"),
  toolVersions: (key: string) => request<ToolVersions>(`/api/tools/${encode(key)}/versions`),
  jobs: () => request<Job[]>("/api/jobs"),
  job: (id: string) => request<Job>(`/api/jobs/${encode(id)}`),
  startJob: (body: JobRequest) => request<Job>("/api/jobs", { method: "POST", body: JSON.stringify(body) }),
};
