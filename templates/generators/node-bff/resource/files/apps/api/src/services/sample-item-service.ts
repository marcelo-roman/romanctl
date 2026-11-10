import { SampleItem } from '../domain/sample-item.js';
import type { SampleItemRepository } from '../repositories/sample-item-repository.js';
import type { Clock } from './clock.js';
import { NotFoundError } from './errors.js';
import { listLimit } from './list-limit.js';

export interface SampleItemInput {
  name: string;
}

export interface ListSampleItemsInput {
  limit?: number;
}

export class SampleItemService {
  constructor(
    private readonly sampleItems: SampleItemRepository,
    private readonly clock: Clock,
  ) {}

  async create(input: SampleItemInput): Promise<SampleItem> {
    const sampleItem = SampleItem.create(input.name, this.clock());
    await this.sampleItems.save(sampleItem);
    return sampleItem;
  }

  list(input: ListSampleItemsInput): Promise<SampleItem[]> {
    return this.sampleItems.latest(listLimit(input.limit));
  }

  async get(id: string): Promise<SampleItem> {
    const sampleItem = await this.sampleItems.findById(id);
    if (!sampleItem) {
      throw new NotFoundError(`Registro ${id} não encontrado.`);
    }
    return sampleItem;
  }

  async rename(id: string, input: SampleItemInput): Promise<SampleItem> {
    const sampleItem = await this.get(id);
    sampleItem.rename(input.name);
    await this.sampleItems.save(sampleItem);
    return sampleItem;
  }

  async delete(id: string): Promise<void> {
    if (!(await this.sampleItems.remove(id))) {
      throw new NotFoundError(`Registro ${id} não encontrado.`);
    }
  }
}
