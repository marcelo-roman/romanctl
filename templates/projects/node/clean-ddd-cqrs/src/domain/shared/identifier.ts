import { DomainError } from './domain-error.js';

const format = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export function newIdentifier(): string {
  return crypto.randomUUID();
}

export function parseIdentifier(value: string): string {
  if (!format.test(value)) {
    throw new DomainError('Identificador inválido.');
  }
  return value.toLowerCase();
}
