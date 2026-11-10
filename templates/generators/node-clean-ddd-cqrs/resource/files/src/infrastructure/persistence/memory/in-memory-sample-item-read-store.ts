import { Inject, Injectable } from '@nestjs/common';
import { SampleItemReadStore } from '../../../application/sample-items/sample-item-read-store.js';
import { type SampleItemView, toSampleItemView } from '../../../application/sample-items/sample-item-view.js';
import { InMemorySampleItemStore } from './in-memory-sample-item-store.js';

@Injectable()
export class InMemorySampleItemReadStore extends SampleItemReadStore {
  constructor(@Inject(InMemorySampleItemStore) private readonly store: InMemorySampleItemStore) {
    super();
  }

  async findById(id: string): Promise<SampleItemView | undefined> {
    const sampleItem = this.store.get(id);
    return sampleItem && toSampleItemView(sampleItem);
  }

  async list(limit: number): Promise<SampleItemView[]> {
    return this.store.latest(limit).map(toSampleItemView);
  }
}
