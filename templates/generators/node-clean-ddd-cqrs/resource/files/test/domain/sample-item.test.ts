import { describe, expect, it } from 'vitest';
import { SampleItem } from '../../src/domain/sample-items/sample-item.js';
import { DomainError } from '../../src/domain/shared/domain-error.js';

const now = new Date('2026-01-01T12:00:00Z');

describe('SampleItem', () => {
  it('requires a name and trims it', () => {
    expect(() => SampleItem.create(' ', now)).toThrow(DomainError);
    expect(SampleItem.create('  Primeiro ', now).name).toBe('Primeiro');
  });

  it('keeps the name when a rename is refused', () => {
    const sampleItem = SampleItem.create('Primeiro', now);
    expect(() => sampleItem.rename('')).toThrow(DomainError);
    sampleItem.rename('Segundo');
    expect(sampleItem.name).toBe('Segundo');
  });
});
