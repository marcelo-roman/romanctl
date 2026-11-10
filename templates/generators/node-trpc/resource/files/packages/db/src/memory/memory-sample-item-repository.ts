import type { SampleItemRecord } from '../sample-items/sample-item-record';
import type { SampleItemRepository } from '../sample-items/sample-item-repository';

export class MemorySampleItemRepository implements SampleItemRepository {
  private readonly sampleItems = new Map<string, SampleItemRecord>();

  async save(sampleItem: SampleItemRecord): Promise<void> {
    this.sampleItems.set(sampleItem.id, structuredClone(sampleItem));
  }

  async findById(id: string): Promise<SampleItemRecord | undefined> {
    const sampleItem = this.sampleItems.get(id);
    return sampleItem && structuredClone(sampleItem);
  }

  async latest(limit: number): Promise<SampleItemRecord[]> {
    return [...this.sampleItems.values()]
      .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
      .slice(0, limit)
      .map((sampleItem) => structuredClone(sampleItem));
  }

  async remove(id: string): Promise<boolean> {
    return this.sampleItems.delete(id);
  }
}
