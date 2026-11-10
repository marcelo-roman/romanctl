import { describe, expect, it } from 'vitest';
import { formatMoney, money, multiplyMoney } from './money';

const spaces = (text: string) => text.replace(/\s/g, ' ');

describe('money', () => {
  it('formats BRL the Brazilian way', () => {
    expect(spaces(formatMoney(money(1234.5)))).toBe('R$ 1.234,50');
    expect(spaces(formatMoney(money(0)))).toBe('R$ 0,00');
  });

  it('formats other currencies with their own symbol', () => {
    expect(spaces(formatMoney(money(10, 'usd')))).toBe('US$ 10,00');
  });

  it('falls back to the code when the currency is unknown to Intl', () => {
    expect(formatMoney({ amount: 2, currency: 'X' })).toBe('X 2.00');
  });

  it('multiplies in cents, without floating point noise', () => {
    expect(multiplyMoney(money(0.1), 3)).toEqual(money(0.3));
    expect(multiplyMoney(money(3.5), 2)).toEqual(money(7));
  });
});
