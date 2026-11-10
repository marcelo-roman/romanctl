import type { SystemSnapshot } from "../api/types";

export interface Usage {
  used: number;
  total: number;
  percent: number;
}

function usage(used: number, total: number): Usage {
  return { used, total, percent: total > 0 ? (used / total) * 100 : 0 };
}

export function cpuPercent(snapshot: SystemSnapshot): number {
  if (snapshot.systemCpuPercent !== null && snapshot.systemCpuPercent !== undefined) return snapshot.systemCpuPercent;
  if (snapshot.cores <= 0) return 0;
  return (snapshot.visibleProcessesCores / snapshot.cores) * 100;
}

export function memoryUsage(snapshot: SystemSnapshot): Usage {
  const { totalBytes, availableBytes } = snapshot.memory;
  return usage(totalBytes - availableBytes, totalBytes);
}

export function diskUsage(snapshot: SystemSnapshot): Usage {
  const { sizeBytes, freeBytes } = snapshot.disk;
  return usage(sizeBytes - freeBytes, sizeBytes);
}

export function level(percent: number): "ok" | "warn" | "danger" {
  if (percent >= 90) return "danger";
  if (percent >= 70) return "warn";
  return "ok";
}
