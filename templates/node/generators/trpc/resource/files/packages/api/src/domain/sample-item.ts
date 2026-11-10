import { DomainError } from './domain-error';

const MAX_NAME_LENGTH = 200;

function validName(name: string): string {
  const trimmed = name.trim();
  if (trimmed === '') {
    throw new DomainError('O nome é obrigatório.');
  }
  if (trimmed.length > MAX_NAME_LENGTH) {
    throw new DomainError(`O nome tem no máximo ${MAX_NAME_LENGTH} caracteres.`);
  }
  return trimmed;
}

export class SampleItem {
  private constructor(
    readonly id: string,
    private label: string,
    readonly createdAt: Date,
  ) {}

  static create(name: string, now: Date): SampleItem {
    return new SampleItem(crypto.randomUUID(), validName(name), now);
  }

  static restore(id: string, name: string, createdAt: Date): SampleItem {
    return new SampleItem(id, name, createdAt);
  }

  get name(): string {
    return this.label;
  }

  rename(name: string): void {
    this.label = validName(name);
  }
}
