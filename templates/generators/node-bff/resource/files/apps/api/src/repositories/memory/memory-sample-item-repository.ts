import { SampleItem } from '../../domain/sample-item.js';
import type { SampleItemRepository } from '../sample-item-repository.js';

function copy(sampleItem: SampleItem): SampleItem {
  return SampleItem.restore(sampleItem.id, sampleItem.name, sampleItem.createdAt);
}

export class MemorySampleItemRepository implements SampleItemRepository {
  private readonly sampleItems = new Map<string, SampleItem>();

  async save(sampleItem: SampleItem): Promise<void> {
    this.sampleItems.set(sampleItem.id, copy(sampleItem));
  }

  async findById(id: string): Promise<SampleItem | undefined> {
    const sampleItem = this.sampleItems.get(id);
    return sampleItem ? copy(sampleItem) : undefined;
  }

  async latest(limit: number): Promise<SampleItem[]> {
    return [...this.sampleItems.values()]
      .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
      .slice(0, limit)
      .map(copy);
  }

  async remove(id: string): Promise<boolean> {
    return this.sampleItems.delete(id);
  }
}
