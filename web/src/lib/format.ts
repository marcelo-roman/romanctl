import type { Lang } from "../api/types";

const units = ["B", "KB", "MB", "GB", "TB"];

export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0 B";
  const exponent = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const value = bytes / 1024 ** exponent;
  return `${value.toFixed(value >= 100 || exponent === 0 ? 0 : 1)} ${units[exponent]}`;
}

export function formatPercent(value: number): string {
  return `${Math.round(value)}%`;
}

export function locale(lang: Lang): string {
  return lang === "en" ? "en-US" : "pt-BR";
}

export function formatTime(iso: string, lang: Lang): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  return date.toLocaleString(locale(lang), { dateStyle: "short", timeStyle: "medium" });
}

export function formatDuration(startIso: string, endIso?: string, now = Date.now()): string {
  const start = new Date(startIso).getTime();
  const end = endIso ? new Date(endIso).getTime() : now;
  const seconds = Math.max(0, Math.round((end - start) / 1000));
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${seconds % 60}s`;
}

export function prettyJson(value: unknown): string {
  return JSON.stringify(value, null, 2);
}
