import { describe, expect, it } from 'vitest';
import { DEFAULT_LIST_LIMIT, MAX_LIST_LIMIT, listLimit } from '../../src/services/list-limit.js';

describe('listLimit', () => {
  it('defaults and clamps the requested limit', () => {
    expect(listLimit(undefined)).toBe(DEFAULT_LIST_LIMIT);
    expect(listLimit(0)).toBe(1);
    expect(listLimit(10)).toBe(10);
    expect(listLimit(1000)).toBe(MAX_LIST_LIMIT);
  });
});
