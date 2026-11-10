import type { SampleItem } from './sample-item.js';
import type { SampleItemId } from './sample-item-id.js';

export abstract class SampleItemRepository {
  abstract save(sampleItem: SampleItem): Promise<void>;
  abstract findById(id: SampleItemId): Promise<SampleItem | undefined>;
  abstract remove(sampleItem: SampleItem): Promise<void>;
}
