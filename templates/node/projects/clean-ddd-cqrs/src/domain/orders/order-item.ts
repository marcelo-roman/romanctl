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
      throw new DomainError('Product name is required.');
    }
    if (!Number.isInteger(quantity)) {
      throw new DomainError('Quantity must be a whole number.');
    }
    if (quantity <= 0) {
      throw new DomainError('Quantity must be greater than zero.');
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
