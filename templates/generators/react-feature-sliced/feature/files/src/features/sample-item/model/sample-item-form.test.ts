import { describe, expect, it } from 'vitest';
import { emptySampleItem, toSampleItemInput, validateSampleItem } from './sample-item-form';

describe('sample item form', () => {
  it('requires a name', () => {
    expect(validateSampleItem(emptySampleItem)).toBe('Informe o nome.');
    expect(validateSampleItem({ name: 'Ana' })).toBeUndefined();
  });

  it('trims the name for the API', () => {
    expect(toSampleItemInput({ name: ' Ana ' })).toEqual({ name: 'Ana' });
  });
});
