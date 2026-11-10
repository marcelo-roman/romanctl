import { SampleItem } from '../../../domain/sample-items/sample-item.js';

export class InMemorySampleItemStore {
  private readonly sampleItems = new Map<string, SampleItem>();

  put(sampleItem: SampleItem): void {
    this.sampleItems.set(sampleItem.id.value, copy(sampleItem));
  }

  get(id: string): SampleItem | undefined {
    const sampleItem = this.sampleItems.get(id);
    return sampleItem && copy(sampleItem);
  }

  delete(id: string): void {
    this.sampleItems.delete(id);
  }

  latest(limit: number): SampleItem[] {
    return [...this.sampleItems.values()]
      .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
      .slice(0, limit)
      .map(copy);
  }
}

function copy(sampleItem: SampleItem): SampleItem {
  return SampleItem.restore(sampleItem.id, sampleItem.name, sampleItem.createdAt);
}
