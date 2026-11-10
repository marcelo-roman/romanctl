import { describe, expect, it } from 'vitest';
import { parseDecimal, parseInteger } from './number';

describe('parseDecimal', () => {
  it('accepts comma and dot as decimal separator', () => {
    expect(parseDecimal('3,50')).toBe(3.5);
    expect(parseDecimal('3.50')).toBe(3.5);
    expect(parseDecimal('1.234,56')).toBe(1234.56);
  });

  it('returns NaN for blank or invalid text', () => {
    expect(parseDecimal(' ')).toBeNaN();
    expect(parseDecimal('abc')).toBeNaN();
  });
});

describe('parseInteger', () => {
  it('rejects fractions', () => {
    expect(parseInteger('2')).toBe(2);
    expect(parseInteger('0')).toBe(0);
    expect(parseInteger('2,5')).toBeNaN();
  });
});
