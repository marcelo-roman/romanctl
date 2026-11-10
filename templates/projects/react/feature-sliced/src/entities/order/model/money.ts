export interface Money {
  amount: number;
  currency: string;
}

export const defaultCurrency = 'BRL';

const formatters = new Map<string, Intl.NumberFormat>();

function toCents(amount: number): number {
  return Math.round(amount * 100);
}

function formatterFor(currency: string): Intl.NumberFormat {
  const cached = formatters.get(currency);
  if (cached) {
    return cached;
  }
  const created = new Intl.NumberFormat('pt-BR', { style: 'currency', currency });
  formatters.set(currency, created);
  return created;
}

export function money(amount: number, currency: string = defaultCurrency): Money {
  return { amount, currency: currency.toUpperCase() };
}

export function multiplyMoney(value: Money, quantity: number): Money {
  return { amount: (toCents(value.amount) * quantity) / 100, currency: value.currency };
}

export function formatMoney(value: Money): string {
  try {
    return formatterFor(value.currency).format(value.amount);
  } catch {
    return `${value.currency} ${value.amount.toFixed(2)}`;
  }
}
