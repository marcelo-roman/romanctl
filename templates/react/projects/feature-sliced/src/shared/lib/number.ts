export function parseDecimal(value: string): number {
  const normalized = value.trim().replace(/\s/g, '');
  if (normalized === '') {
    return Number.NaN;
  }
  const decimal = normalized.includes(',') ? normalized.replace(/\./g, '').replace(',', '.') : normalized;
  return Number(decimal);
}

export function parseInteger(value: string): number {
  const parsed = parseDecimal(value);
  return Number.isInteger(parsed) ? parsed : Number.NaN;
}
