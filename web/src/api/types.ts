export type Lang = "pt" | "en";

export interface PlatformInfo {
  os: string;
  arch: string;
  wsl?: boolean;
  proot?: boolean;
  manager?: string;
  libc?: string;
}

export interface Info {
  version: string;
  lang: Lang;
  platform: PlatformInfo;
  projectsDir: string;
  templates: { installed: boolean; ref: string; path: string };
}

export interface Battery {
  source: string;
  percent: number;
  status: string;
  powerW?: number;
  hoursLeft?: number;
}

export interface Process {
  pid: number;
  name: string;
  cpuPercentOfOneCore: number;
  rssBytes: number;
}

export interface SystemSnapshot {
  platform: PlatformInfo;
  cores: number;
  systemCpuPercent: number | null;
  visibleProcessesCores: number;
  visibleProcessesRssBytes: number;
  memory: { totalBytes: number; availableBytes: number; swapTotalBytes: number; swapFreeBytes: number };
  disk: { path: string; sizeBytes: number; freeBytes: number };
  battery?: Battery;
  batteryError?: string;
  top: Process[];
}

export interface Generator {
  kind: string;
  description: string;
  aliases?: string[];
  needsIn: boolean;
}

export interface Manifest {
  name: string;
  stack: string;
  template: string;
  arch: string[] | null;
  services: string[] | null;
  url: string;
}

export interface ProjectUp {
  pid: number;
  alive: boolean;
  url: string;
  services: string[] | null;
}

export interface Project {
  name: string;
  path: string;
  type: string;
  modified: string;
  url: string;
  manifest?: Manifest;
  up?: ProjectUp;
  target: string;
  generators: Generator[];
}

export interface LogLines {
  lines: string[];
  running?: boolean;
  path?: string;
}

export interface Template {
  id: string;
  stack: string;
  arch: string[];
  description: string;
  services: string[] | null;
  requires: string[] | null;
  url: string;
}

export interface SimpleType {
  type: string;
  description: string;
}

export interface Catalog {
  stacks: string[];
  arch: string[];
  services: string[];
  templates: Template[];
  simpleTypes: SimpleType[];
  generators: Record<string, Generator[]>;
}

export interface Service {
  key: string;
  name: string;
  installed: boolean;
  running: boolean;
  engine: boolean;
  unsupported: string;
  url: string;
  wanted: boolean;
}

export interface ToolGroup {
  key: string;
  title: string;
}

export interface Tool {
  key: string;
  name: string;
  group: string;
  description: string;
  installed: boolean;
  version: string;
  latest?: string;
  unavailable?: string;
}

export interface ToolList {
  groups: ToolGroup[];
  tools: Tool[];
}

export interface ToolVersion {
  id: string;
  label: string;
  note?: string;
}

export interface ToolVersions {
  versions: ToolVersion[];
  error?: string;
}

export type JobStatus = "running" | "succeeded" | "failed";

export interface Job {
  id: string;
  action: string;
  title: string;
  args: string[];
  status: JobStatus;
  started: string;
  finished?: string;
  lines: string[];
  result?: unknown;
  error?: string;
}

export type JobRequest =
  | { action: "project.create"; name: string; type: string; git?: boolean }
  | { action: "project.create"; name: string; stack: string; arch: string[]; with: string[]; git?: boolean }
  | { action: "project.up" | "project.down" | "project.remove"; name: string }
  | { action: "project.generate"; name: string; kind: string; entity: string; in?: string; plural?: string }
  | { action: "service.install" | "service.start" | "service.stop" | "service.restart"; key: string }
  | { action: "tool.install"; key: string; version?: string }
  | { action: "tool.uninstall"; key: string }
  | { action: "templates.update" };
