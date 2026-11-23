const labels: Record<string, string> = {
  Draft: 'Draft',
  Placed: 'Placed',
};

export function statusLabel(status: string): string {
  return labels[status] ?? status;
}
