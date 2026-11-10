import { useEffect, useRef } from "react";

interface LogViewerProps {
  lines: string[];
  emptyText: string;
  label: string;
}

export function LogViewer({ lines, emptyText, label }: LogViewerProps) {
  const ref = useRef<HTMLPreElement>(null);

  useEffect(() => {
    const element = ref.current;
    if (!element) return;
    element.scrollTop = element.scrollHeight;
  }, [lines]);

  if (lines.length === 0) return <p className="muted">{emptyText}</p>;
  return (
    <pre ref={ref} className="log" aria-label={label} tabIndex={0}>
      {lines.join("\n")}
    </pre>
  );
}
