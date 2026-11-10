import { randomUUID } from 'node:crypto';
import { DomainError } from './domain-error.js';

export const SAMPLE_ITEM_NAME_MAX_LENGTH = 200;

function requiredName(name: string): string {
  const trimmed = name.trim();
  if (trimmed === '') {
    throw new DomainError('O nome é obrigatório.');
  }
  if (trimmed.length > SAMPLE_ITEM_NAME_MAX_LENGTH) {
    throw new DomainError(`O nome deve ter no máximo ${SAMPLE_ITEM_NAME_MAX_LENGTH} caracteres.`);
  }
  return trimmed;
}

export class SampleItem {
  private constructor(
    readonly id: string,
    private currentName: string,
    readonly createdAt: Date,
  ) {}

  static create(name: string, now: Date): SampleItem {
    return new SampleItem(randomUUID(), requiredName(name), now);
  }

  static restore(id: string, name: string, createdAt: Date): SampleItem {
    return new SampleItem(id, name, createdAt);
  }

  get name(): string {
    return this.currentName;
  }

  rename(name: string): void {
    this.currentName = requiredName(name);
  }
}
