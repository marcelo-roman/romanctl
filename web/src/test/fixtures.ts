import type { Catalog, Info, Job, Project, Service, SystemSnapshot, ToolList } from "../api/types";

export const testToken = "test-token-123";

export const info: Info = {
  version: "9.9.9",
  lang: "pt",
  platform: { os: "linux", arch: "amd64", wsl: true },
  projectsDir: "/tmp/proj",
  templates: { installed: true, ref: "v1.2.0", path: "/tmp/templates" },
};

export const system: SystemSnapshot = {
  platform: { os: "linux", arch: "amd64" },
  cores: 8,
  systemCpuPercent: 25,
  visibleProcessesCores: 1,
  visibleProcessesRssBytes: 1024,
  memory: { totalBytes: 16 * 1024 ** 3, availableBytes: 8 * 1024 ** 3, swapTotalBytes: 0, swapFreeBytes: 0 },
  disk: { path: "/home", sizeBytes: 100 * 1024 ** 3, freeBytes: 40 * 1024 ** 3 },
  top: [{ pid: 42, name: "node", cpuPercentOfOneCore: 12, rssBytes: 1024 ** 2 }],
};

export const projects: Project[] = [
  {
    name: "loja",
    path: "/tmp/proj/loja",
    type: "go/hexagonal-ddd",
    modified: "2026-11-10T12:00:00Z",
    url: "http://localhost:8090",
    manifest: { name: "loja", stack: "go", template: "hexagonal-ddd", arch: ["ddd", "hexagonal"], services: [], url: "http://localhost:8090" },
    target: "go/hexagonal-ddd",
    generators: [{ kind: "resource", description: "CRUD completo", aliases: ["crud"], needsIn: false }],
  },
  {
    name: "site",
    path: "/tmp/proj/site",
    type: "react",
    modified: "2026-11-09T12:00:00Z",
    url: "http://localhost:5173",
    up: { pid: 10, alive: true, url: "http://localhost:5173", services: [] },
    target: "",
    generators: [],
  },
];

export const catalog: Catalog = {
  stacks: ["dotnet", "go"],
  arch: ["clean", "cqrs", "ddd", "hexagonal"],
  services: ["postgres", "redis"],
  templates: [
    {
      id: "clean-ddd-cqrs",
      stack: "dotnet",
      arch: ["clean", "cqrs", "ddd"],
      description: "API .NET com Clean Architecture",
      services: ["postgres"],
      requires: ["dotnet"],
      url: "http://localhost:5080",
    },
    {
      id: "hexagonal-ddd",
      stack: "go",
      arch: ["ddd", "hexagonal"],
      description: "API Go hexagonal com DDD",
      services: ["postgres", "redis"],
      requires: ["go"],
      url: "http://localhost:8090",
    },
  ],
  simpleTypes: [
    { type: "node", description: "Node.js (básico)" },
    { type: "go", description: "Go (básico)" },
  ],
  generators: {
    "go/hexagonal-ddd": [{ kind: "resource", description: "CRUD completo", aliases: ["crud"], needsIn: false }],
  },
};

export const services: Service[] = [
  { key: "docker", name: "Docker", installed: true, running: true, engine: true, unsupported: "", url: "unix:///var/run/docker.sock", wanted: false },
  { key: "postgres", name: "PostgreSQL", installed: true, running: false, engine: false, unsupported: "", url: "postgres://127.0.0.1:5432", wanted: false },
];

export const tools: ToolList = {
  groups: [
    { key: "sdk", title: "Linguagens" },
    { key: "k8s", title: "Kubernetes" },
  ],
  tools: [
    { key: "node", name: "Node.js", group: "sdk", description: "Versão LTS", installed: true, version: "v24.1.0", latest: "v24.2.0" },
    { key: "go", name: "Go", group: "sdk", description: "Última estável", installed: false, version: "" },
    { key: "kubectl", name: "kubectl", group: "k8s", description: "CLI do Kubernetes", installed: true, version: "1.31.0", latest: "1.31.0" },
    { key: "helm", name: "Helm", group: "k8s", description: "Gerenciador de pacotes", installed: false, version: "", unavailable: "só existe para macOS" },
  ],
};

export function job(overrides: Partial<Job> = {}): Job {
  return {
    id: "1",
    action: "project.up",
    title: "Subir loja",
    args: ["up", "loja", "-d"],
    status: "running",
    started: "2026-11-10T12:00:00Z",
    lines: [],
    ...overrides,
  };
}
