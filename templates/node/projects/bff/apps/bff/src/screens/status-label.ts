const labels: Record<string, string> = {
  Draft: 'Rascunho',
  Placed: 'Realizado',
};

export function statusLabel(status: string): string {
  return labels[status] ?? status;
}
