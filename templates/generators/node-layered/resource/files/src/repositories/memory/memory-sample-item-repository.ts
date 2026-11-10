import type { SampleItem } from '../../models/sample-item.js';
import type { SampleItemRepository } from '../sample-item-repository.js';

export class MemorySampleItemRepository implements SampleItemRepository {
  private readonly sampleItems = new Map<string, SampleItem>();

  async add(sampleItem: SampleItem): Promise<void> {
    this.sampleItems.set(sampleItem.id, { ...sampleItem });
  }

  async findById(id: string): Promise<SampleItem | undefined> {
    const sampleItem = this.sampleItems.get(id);
    return sampleItem ? { ...sampleItem } : undefined;
  }

  async list(limit: number): Promise<SampleItem[]> {
    return [...this.sampleItems.values()]
      .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
      .slice(0, limit)
      .map((sampleItem) => ({ ...sampleItem }));
  }

  async update(sampleItem: SampleItem): Promise<void> {
    this.sampleItems.set(sampleItem.id, { ...sampleItem });
  }

  async remove(id: string): Promise<boolean> {
    return this.sampleItems.delete(id);
  }
}
