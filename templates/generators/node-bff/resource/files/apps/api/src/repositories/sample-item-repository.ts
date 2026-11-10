import type { SampleItem } from '../domain/sample-item.js';

export interface SampleItemRepository {
  save(sampleItem: SampleItem): Promise<void>;
  findById(id: string): Promise<SampleItem | undefined>;
  latest(limit: number): Promise<SampleItem[]>;
  remove(id: string): Promise<boolean>;
}
