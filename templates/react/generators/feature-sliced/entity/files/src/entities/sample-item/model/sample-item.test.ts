import { describe, expect, it } from 'vitest';
import { newestFirst, sampleItemLabel, type SampleItem } from './sample-item';

const sampleItem = (id: string, name: string, day: number): SampleItem => ({
  id,
  name,
  createdAt: new Date(Date.UTC(2026, 0, day)),
});

describe('sample item', () => {
  it('labels it by name, with a fallback for a blank one', () => {
    expect(sampleItemLabel(sampleItem('1', ' Ana ', 1))).toBe('Ana');
    expect(sampleItemLabel(sampleItem('2', ' ', 1))).toBe('Sem nome');
  });

  it('sorts the newest first without changing the original list', () => {
    const list = [sampleItem('old', 'a', 1), sampleItem('new', 'b', 2)];

    expect(newestFirst(list).map((item) => item.id)).toEqual(['new', 'old']);
    expect(list.map((item) => item.id)).toEqual(['old', 'new']);
  });
});
