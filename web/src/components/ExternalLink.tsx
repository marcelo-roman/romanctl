import { ExternalLink as ExternalIcon } from "lucide-react";

function isHttp(url: string): boolean {
  return /^https?:\/\//.test(url);
}

export function ExternalLink({ url }: { url: string }) {
  if (!url) return <span className="muted">—</span>;
  if (!isHttp(url)) return <code className="mono">{url}</code>;
  return (
    <a className="link mono" href={url} target="_blank" rel="noreferrer">
      {url}
      <ExternalIcon size={12} aria-hidden="true" />
    </a>
  );
}
