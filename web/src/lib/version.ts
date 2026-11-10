const versionPattern = /\d+(?:\.\d+)*/;

export function extractVersion(text: string | undefined): number[] | null {
  if (!text) return null;
  const match = versionPattern.exec(text);
  if (!match) return null;
  return match[0].split(".").map(Number);
}

export function isNewer(latest: string | undefined, current: string | undefined): boolean {
  const a = extractVersion(latest);
  const b = extractVersion(current);
  if (!a || !b) return false;
  const length = Math.max(a.length, b.length);
  for (let i = 0; i < length; i++) {
    const left = a[i] ?? 0;
    const right = b[i] ?? 0;
    if (left !== right) return left > right;
  }
  return false;
}

export function shortVersion(text: string | undefined): string {
  const parts = extractVersion(text);
  if (!parts) return text ?? "";
  return parts.join(".");
}
