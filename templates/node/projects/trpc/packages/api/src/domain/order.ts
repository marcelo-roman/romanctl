import { DomainError, InvalidStateError } from './domain-error';
import { DEFAULT_CURRENCY, Money } from './money';
import { OrderItem } from './order-item';
import { OrderStatus } from './order-status';

export class Order {
  private constructor(
    readonly id: string,
    readonly customerName: string,
    private state: OrderStatus,
    readonly createdAt: Date,
    private readonly lines: OrderItem[],
  ) {}

  static create(customerName: string, now: Date): Order {
    if (customerName.trim() === '') {
      throw new DomainError('O nome do cliente é obrigatório.');
    }
    return new Order(crypto.randomUUID(), customerName.trim(), OrderStatus.Draft, now, []);
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

  get total(): Money {
    const currency = this.lines[0]?.unitPrice.currency ?? DEFAULT_CURRENCY;
    return this.lines.reduce((total, item) => total.add(item.subtotal), Money.fromCents(0, currency));
  }

  addItem(productName: string, quantity: number, unitPrice: Money): void {
    if (this.state !== OrderStatus.Draft) {
      throw new InvalidStateError('Um pedido já realizado não pode ser alterado.');
    }
    const first = this.lines[0];
    if (first && first.unitPrice.currency !== unitPrice.currency) {
      throw new DomainError('Todos os itens do pedido devem usar a mesma moeda.');
    }
    this.lines.push(OrderItem.create(productName, quantity, unitPrice));
  }

  place(): void {
    if (this.state !== OrderStatus.Draft) {
      throw new InvalidStateError('O pedido já foi realizado.');
    }
    if (this.lines.length === 0) {
      throw new DomainError('Um pedido sem itens não pode ser realizado.');
    }
    this.state = OrderStatus.Placed;
  }
}
