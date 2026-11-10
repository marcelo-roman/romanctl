import type { SampleItem } from '../../domain/sample-items/sample-item.js';

export interface SampleItemView {
  id: string;
  name: string;
  createdAt: string;
}

export function toSampleItemView(sampleItem: SampleItem): SampleItemView {
  return {
    id: sampleItem.id.value,
    name: sampleItem.name,
    createdAt: sampleItem.createdAt.toISOString(),
  };
}
