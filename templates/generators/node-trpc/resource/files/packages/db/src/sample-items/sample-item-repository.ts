import type { SampleItemRecord } from './sample-item-record';

export interface SampleItemRepository {
  save(sampleItem: SampleItemRecord): Promise<void>;
  findById(id: string): Promise<SampleItemRecord | undefined>;
  latest(limit: number): Promise<SampleItemRecord[]>;
  remove(id: string): Promise<boolean>;
}
