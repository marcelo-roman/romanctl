import { randomUUID } from 'node:crypto';
import { DomainError } from './domain-error.js';
import { DEFAULT_CURRENCY, Money } from './money.js';
import { OrderItem } from './order-item.js';
import { OrderStatus } from './order-status.js';

export const CUSTOMER_NAME_MAX_LENGTH = 200;

export class Order {
  private constructor(
    readonly id: string,
    readonly customerName: string,
    private state: OrderStatus,
    readonly createdAt: Date,
    private readonly lines: OrderItem[],
  ) {}

  static create(customerName: string, now: Date): Order {
    const name = customerName.trim();
    if (name === '') {
      throw new DomainError('O nome do cliente é obrigatório.');
    }
    if (name.length > CUSTOMER_NAME_MAX_LENGTH) {
      throw new DomainError(`O nome do cliente deve ter no máximo ${CUSTOMER_NAME_MAX_LENGTH} caracteres.`);
    }
    return new Order(randomUUID(), name, OrderStatus.Draft, now, []);
  }

  static restore(id: string, customerName: string, status: OrderStatus, createdAt: Date, items: OrderItem[]): Order {
    return new Order(id, customerName, status, createdAt, [...items]);
  }

  get status(): OrderStatus {
    return this.state;
  }

  get items(): readonly OrderItem[] {
    return [...this.lines];
  }

  get currency(): string {
    return this.lines[0]?.unitPrice.currency ?? DEFAULT_CURRENCY;
  }

  get total(): Money {
    return this.lines.reduce((total, item) => total.add(item.subtotal), Money.fromCents(0, this.currency));
  }

  get canAddItems(): boolean {
    return this.state === OrderStatus.Draft;
  }

  get canPlace(): boolean {
    return this.state === OrderStatus.Draft && this.lines.length > 0;
  }

  addItem(productName: string, quantity: number, unitPrice: Money): OrderItem {
    if (!this.canAddItems) {
      throw new DomainError('Um pedido já realizado não pode ser alterado.');
    }
    if (this.lines.length > 0 && this.currency !== unitPrice.currency) {
      throw new DomainError('Todos os itens do pedido devem usar a mesma moeda.');
    }
    const item = OrderItem.create(productName, quantity, unitPrice);
    this.lines.push(item);
    return item;
  }

  place(): void {
    if (this.state !== OrderStatus.Draft) {
      throw new DomainError('O pedido já foi realizado.');
    }
    if (this.lines.length === 0) {
      throw new DomainError('Um pedido sem itens não pode ser realizado.');
    }
    this.state = OrderStatus.Placed;
  }
}
