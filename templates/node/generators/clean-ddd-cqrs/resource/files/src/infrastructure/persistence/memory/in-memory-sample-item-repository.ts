import { Inject, Injectable } from '@nestjs/common';
import type { SampleItem } from '../../../domain/sample-items/sample-item.js';
import type { SampleItemId } from '../../../domain/sample-items/sample-item-id.js';
import { SampleItemRepository } from '../../../domain/sample-items/sample-item-repository.js';
import { InMemorySampleItemStore } from './in-memory-sample-item-store.js';

@Injectable()
export class InMemorySampleItemRepository extends SampleItemRepository {
  constructor(@Inject(InMemorySampleItemStore) private readonly store: InMemorySampleItemStore) {
    super();
  }

  async save(sampleItem: SampleItem): Promise<void> {
    this.store.put(sampleItem);
  }

  async findById(id: SampleItemId): Promise<SampleItem | undefined> {
    return this.store.get(id.value);
  }

  async remove(sampleItem: SampleItem): Promise<void> {
    this.store.delete(sampleItem.id.value);
  }
}
