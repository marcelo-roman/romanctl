import type { SampleItemView } from './sample-item-view.js';

export abstract class SampleItemReadStore {
  abstract findById(id: string): Promise<SampleItemView | undefined>;
  abstract list(limit: number): Promise<SampleItemView[]>;
}
