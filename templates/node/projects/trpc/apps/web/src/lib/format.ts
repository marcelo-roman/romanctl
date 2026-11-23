const statusLabels: Record<string, string> = {
  Draft: 'Draft',
  Placed: 'Placed',
};

export function formatMoney(amount: number, currency: string): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(amount);
}

export function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('en-US', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(iso));
}

export function statusLabel(status: string): string {
  return statusLabels[status] ?? status;
}

export function parseDecimal(text: string): number {
  const normalized = text.trim().replace(',', '.');
  return normalized === '' ? Number.NaN : Number(normalized);
}
