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
      throw new DomainError('O nome do produto é obrigatório.');
    }
    if (!Number.isInteger(quantity)) {
      throw new DomainError('A quantidade deve ser um número inteiro.');
    }
    if (quantity <= 0) {
      throw new DomainError('A quantidade deve ser maior que zero.');
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
