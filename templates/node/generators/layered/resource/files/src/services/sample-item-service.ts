import { randomUUID } from 'node:crypto';
import { NAME_MAX_LENGTH, type SampleItem } from '../models/sample-item.js';
import type { SampleItemRepository } from '../repositories/sample-item-repository.js';
import type { Clock } from './clock.js';
import { NotFoundError, ValidationError } from './errors.js';
import { listLimit } from './list-limit.js';

export interface SampleItemInput {
  name: string;
}

export interface ListSampleItemsInput {
  limit?: number;
}

function requiredName(name: string): string {
  const trimmed = name.trim();
  if (trimmed === '') {
    throw new ValidationError('O nome é obrigatório.');
  }
  if (trimmed.length > NAME_MAX_LENGTH) {
    throw new ValidationError(`O nome deve ter no máximo ${NAME_MAX_LENGTH} caracteres.`);
  }
  return trimmed;
}

export class SampleItemService {
  constructor(
    private readonly sampleItems: SampleItemRepository,
    private readonly clock: Clock,
  ) {}

  async create(input: SampleItemInput): Promise<SampleItem> {
    const sampleItem: SampleItem = { id: randomUUID(), name: requiredName(input.name), createdAt: this.clock() };
    await this.sampleItems.add(sampleItem);
    return sampleItem;
  }

  list(input: ListSampleItemsInput): Promise<SampleItem[]> {
    return this.sampleItems.list(listLimit(input.limit));
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
    const renamed: SampleItem = { ...sampleItem, name: requiredName(input.name) };
    await this.sampleItems.update(renamed);
    return renamed;
  }

  async delete(id: string): Promise<void> {
    const removed = await this.sampleItems.remove(id);
    if (!removed) {
      throw new NotFoundError(`Registro ${id} não encontrado.`);
    }
  }
}
