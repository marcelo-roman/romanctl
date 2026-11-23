import { DomainError } from './domain-error';
import type { Money } from './money';

export class OrderItem {
  private constructor(
    readonly id: string,
    readonly productName: string,
    readonly quantity: number,
    readonly unitPrice: Money,
  ) {}

  static create(productName: string, quantity: number, unitPrice: Money): OrderItem {
    if (productName.trim() === '') {
      throw new DomainError('The product name is required.');
    }
    if (!Number.isInteger(quantity)) {
      throw new DomainError('The quantity must be a whole number.');
    }
    if (quantity <= 0) {
      throw new DomainError('The quantity must be greater than zero.');
    }
    return new OrderItem(crypto.randomUUID(), productName.trim(), quantity, unitPrice);
  }

  static restore(id: string, productName: string, quantity: number, unitPrice: Money): OrderItem {
    return new OrderItem(id, productName, quantity, unitPrice);
  }

  get subtotal(): Money {
    return this.unitPrice.multiply(this.quantity);
  }
}
