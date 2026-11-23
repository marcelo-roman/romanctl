import type { SampleItemInput } from '../api/submit-sample-item';

export interface SampleItemValues {
  name: string;
}

export const emptySampleItem: SampleItemValues = { name: '' };

export function validateSampleItem(values: SampleItemValues): string | undefined {
  if (values.name.trim() === '') {
    return 'Enter the name.';
  }
  return undefined;
}

export function toSampleItemInput(values: SampleItemValues): SampleItemInput {
  return { name: values.name.trim() };
}
