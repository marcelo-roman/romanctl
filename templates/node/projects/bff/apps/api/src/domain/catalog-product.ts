import { randomUUID } from 'node:crypto';
import { DomainError } from './domain-error.js';
import type { Money } from './money.js';

export const CATALOG_NAME_MAX_LENGTH = 200;

export class CatalogProduct {
  private constructor(
    readonly id: string,
    readonly name: string,
    readonly price: Money,
    readonly createdAt: Date,
  ) {}

  static create(name: string, price: Money, now: Date): CatalogProduct {
    const trimmed = name.trim();
    if (trimmed === '') {
      throw new DomainError('O nome do produto é obrigatório.');
    }
    if (trimmed.length > CATALOG_NAME_MAX_LENGTH) {
      throw new DomainError(`O nome do produto deve ter no máximo ${CATALOG_NAME_MAX_LENGTH} caracteres.`);
    }
    return new CatalogProduct(randomUUID(), trimmed, price, now);
  }

  static restore(id: string, name: string, price: Money, createdAt: Date): CatalogProduct {
    return new CatalogProduct(id, name, price, createdAt);
  }
}
