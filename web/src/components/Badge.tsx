import type { ReactNode } from "react";

export type Tone = "neutral" | "ok" | "warn" | "danger" | "info";

export function Badge({ tone = "neutral", children, title }: { tone?: Tone; children: ReactNode; title?: string }) {
  return (
    <span className={`badge badge-${tone}`} title={title}>
      {children}
    </span>
  );
}

export function StatusDot({ on, label }: { on: boolean; label: string }) {
  return (
    <span className="status">
      <span className={on ? "dot dot-on" : "dot"} aria-hidden="true" />
      {label}
    </span>
  );
}
