import { postJson } from '@/shared/api';

export interface SampleItemInput {
  name: string;
}

export function submitSampleItem(input: SampleItemInput): Promise<void> {
  return postJson<void>('/sample-item', input);
}
