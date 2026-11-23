import { describe, expect, it } from 'vitest';
import { SampleItem } from '../../src/domain/sample-items/sample-item.js';
import { DomainError } from '../../src/domain/shared/domain-error.js';

const now = new Date('2026-01-01T12:00:00Z');

describe('SampleItem', () => {
  it('requires a name and trims it', () => {
    expect(() => SampleItem.create(' ', now)).toThrow(DomainError);
    expect(SampleItem.create('  First ', now).name).toBe('First');
  });

  it('keeps the name when a rename is refused', () => {
    const sampleItem = SampleItem.create('First', now);
    expect(() => sampleItem.rename('')).toThrow(DomainError);
    sampleItem.rename('Second');
    expect(sampleItem.name).toBe('Second');
  });
});
