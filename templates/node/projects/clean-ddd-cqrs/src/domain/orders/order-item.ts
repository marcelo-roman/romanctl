import { DomainError } from '../shared/domain-error.js';
import { newIdentifier } from '../shared/identifier.js';
import type { Money } from './money.js';

export class OrderItem {
  private constructor(
    readonly id: string,
    readonly productName: string,
    readonly quantity: number,
    readonly unitPrice: Money,
  ) {}

  static create(productName: string, quantity: number, unitPrice: Money): OrderItem {
    if (productName.trim() === '') {
      throw new DomainError('O nome do produto é obrigatório.');
    }
    if (!Number.isInteger(quantity)) {
      throw new DomainError('A quantidade deve ser um número inteiro.');
    }
    if (quantity <= 0) {
      throw new DomainError('A quantidade deve ser maior que zero.');
    }
    return new OrderItem(newIdentifier(), productName.trim(), quantity, unitPrice);
  }

  static restore(id: string, productName: string, quantity: number, unitPrice: Money): OrderItem {
    return new OrderItem(id, productName, quantity, unitPrice);
  }

  get subtotal(): Money {
    return this.unitPrice.multiply(this.quantity);
  }
}
