import { randomUUID } from 'node:crypto';
import { DomainError } from './domain-error.js';
import type { Money } from './money.js';

export const PRODUCT_NAME_MAX_LENGTH = 200;

export class OrderItem {
  private constructor(
    readonly id: string,
    readonly productName: string,
    readonly quantity: number,
    readonly unitPrice: Money,
  ) {}

  static create(productName: string, quantity: number, unitPrice: Money): OrderItem {
    const name = productName.trim();
    if (name === '') {
      throw new DomainError('O nome do produto é obrigatório.');
    }
    if (name.length > PRODUCT_NAME_MAX_LENGTH) {
      throw new DomainError(`O nome do produto deve ter no máximo ${PRODUCT_NAME_MAX_LENGTH} caracteres.`);
    }
    if (!Number.isInteger(quantity)) {
      throw new DomainError('A quantidade deve ser um número inteiro.');
    }
    if (quantity <= 0) {
      throw new DomainError('A quantidade deve ser maior que zero.');
    }
    return new OrderItem(randomUUID(), name, quantity, unitPrice);
  }

  static restore(id: string, productName: string, quantity: number, unitPrice: Money): OrderItem {
    return new OrderItem(id, productName, quantity, unitPrice);
  }

  get subtotal(): Money {
    return this.unitPrice.multiply(this.quantity);
  }
}
