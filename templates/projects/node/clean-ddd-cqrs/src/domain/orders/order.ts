import { AggregateRoot } from '../shared/aggregate-root.js';
import { DomainError } from '../shared/domain-error.js';
import { DEFAULT_CURRENCY, Money } from './money.js';
import { OrderId } from './order-id.js';
import { OrderItem } from './order-item.js';
import { OrderPlaced } from './order-placed.js';
import { OrderStatus } from './order-status.js';

export class Order extends AggregateRoot {
  private constructor(
    readonly id: OrderId,
    readonly customerName: string,
    private state: OrderStatus,
    readonly createdAt: Date,
    private readonly lines: OrderItem[],
  ) {
    super();
  }

  static create(customerName: string, now: Date): Order {
    if (customerName.trim() === '') {
      throw new DomainError('O nome do cliente é obrigatório.');
    }
    return new Order(OrderId.create(), customerName.trim(), OrderStatus.Draft, now, []);
  }

  static restore(id: OrderId, customerName: string, status: OrderStatus, createdAt: Date, items: OrderItem[]): Order {
    return new Order(id, customerName, status, createdAt, [...items]);
  }

  get status(): OrderStatus {
    return this.state;
  }

  get items(): readonly OrderItem[] {
    return [...this.lines];
  }

  get total(): Money {
    const currency = this.lines[0]?.unitPrice.currency ?? DEFAULT_CURRENCY;
    return this.lines.reduce((total, item) => total.add(item.subtotal), Money.fromCents(0, currency));
  }

  addItem(productName: string, quantity: number, unitPrice: Money): void {
    if (this.state !== OrderStatus.Draft) {
      throw new DomainError('Um pedido já realizado não pode ser alterado.');
    }
    if (this.lines.length > 0 && this.lines[0]?.unitPrice.currency !== unitPrice.currency) {
      throw new DomainError('Todos os itens do pedido devem usar a mesma moeda.');
    }
    this.lines.push(OrderItem.create(productName, quantity, unitPrice));
  }

  place(now: Date): void {
    if (this.state !== OrderStatus.Draft) {
      throw new DomainError('O pedido já foi realizado.');
    }
    if (this.lines.length === 0) {
      throw new DomainError('Um pedido sem itens não pode ser realizado.');
    }
    this.state = OrderStatus.Placed;
    const total = this.total;
    this.raise(new OrderPlaced(this.id.value, total.amount, total.currency, now));
  }
}
