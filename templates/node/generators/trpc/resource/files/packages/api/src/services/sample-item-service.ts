import type { SampleItemRecord, SampleItemRepository } from '@sample-app/db';
import { SampleItem } from '../domain/sample-item';
import type { Clock } from './clock';
import { listLimit } from './list-limit';
import { NotFoundError } from './not-found-error';

export interface SampleItemView {
  id: string;
  name: string;
  createdAt: string;
}

function toRecord(sampleItem: SampleItem): SampleItemRecord {
  return { id: sampleItem.id, name: sampleItem.name, createdAt: sampleItem.createdAt };
}

function toView(sampleItem: SampleItem): SampleItemView {
  return { id: sampleItem.id, name: sampleItem.name, createdAt: sampleItem.createdAt.toISOString() };
}

function fromRecord(record: SampleItemRecord): SampleItem {
  return SampleItem.restore(record.id, record.name, record.createdAt);
}

export class SampleItemService {
  constructor(
    private readonly sampleItems: SampleItemRepository,
    private readonly clock: Clock,
  ) {}

  async create(name: string): Promise<SampleItemView> {
    const sampleItem = SampleItem.create(name, this.clock.now());
    await this.sampleItems.save(toRecord(sampleItem));
    return toView(sampleItem);
  }

  async rename(id: string, name: string): Promise<SampleItemView> {
    const sampleItem = await this.load(id);
    sampleItem.rename(name);
    await this.sampleItems.save(toRecord(sampleItem));
    return toView(sampleItem);
  }

  async byId(id: string): Promise<SampleItemView> {
    return toView(await this.load(id));
  }

  async list(limit?: number): Promise<SampleItemView[]> {
    const records = await this.sampleItems.latest(listLimit(limit));
    return records.map((record) => toView(fromRecord(record)));
  }

  async remove(id: string): Promise<void> {
    if (!(await this.sampleItems.remove(id))) {
      throw new NotFoundError('Registro não encontrado.');
    }
  }

  private async load(id: string): Promise<SampleItem> {
    const record = await this.sampleItems.findById(id);
    if (!record) {
      throw new NotFoundError('Registro não encontrado.');
    }
    return fromRecord(record);
  }
}
