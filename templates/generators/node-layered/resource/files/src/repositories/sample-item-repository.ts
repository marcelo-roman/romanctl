import type { SampleItem } from '../models/sample-item.js';

export interface SampleItemRepository {
  add(sampleItem: SampleItem): Promise<void>;
  findById(id: string): Promise<SampleItem | undefined>;
  list(limit: number): Promise<SampleItem[]>;
  update(sampleItem: SampleItem): Promise<void>;
  remove(id: string): Promise<boolean>;
}
