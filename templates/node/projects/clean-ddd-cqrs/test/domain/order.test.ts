import { describe, expect, it } from 'vitest';
import { Money } from '../../src/domain/orders/money.js';
import { Order } from '../../src/domain/orders/order.js';
import { OrderId } from '../../src/domain/orders/order-id.js';
import { OrderPlaced } from '../../src/domain/orders/order-placed.js';
import { OrderStatus } from '../../src/domain/orders/order-status.js';
import { DomainError } from '../../src/domain/shared/domain-error.js';

const now = new Date('2026-01-01T12:00:00Z');
const usd = (cents: number) => Money.fromCents(cents, 'usd');

describe('Order', () => {
  it('requires a customer name', () => {
    expect(() => Order.create('  ', now)).toThrow(DomainError);
  });

  it('starts as a draft with the trimmed customer name', () => {
    const order = Order.create('  Ana ', now);
    expect(order.customerName).toBe('Ana');
    expect(order.status).toBe(OrderStatus.Draft);
    expect(order.total).toEqual(Money.fromCents(0, 'USD'));
  });

  it('sums the items into the total', () => {
    const order = Order.create('Ana', now);
    order.addItem('Pen', 2, usd(350));
    order.addItem('Notebook', 1, usd(1200));
    expect(order.total.cents).toBe(1900);
    expect(order.total.currency).toBe('USD');
  });

  it('requires a product name and a positive whole quantity', () => {
    const order = Order.create('Ana', now);
    expect(() => order.addItem(' ', 1, usd(100))).toThrow(DomainError);
    expect(() => order.addItem('Pen', 0, usd(100))).toThrow(DomainError);
    expect(() => order.addItem('Pen', 1.5, usd(100))).toThrow(DomainError);
  });

  it('keeps every item in the same currency', () => {
    const order = Order.create('Ana', now);
    order.addItem('Pen', 1, usd(100));
    expect(() => order.addItem('Pen', 1, Money.fromCents(100, 'EUR'))).toThrow('same currency');
  });

  it('raises OrderPlaced once and then refuses changes', () => {
    const order = Order.create('Ana', now);
    expect(() => order.place(now)).toThrow('without items');
    order.addItem('Pen', 1, usd(100));
    order.place(now);

    const events = order.pullEvents();
    expect(events).toEqual([new OrderPlaced(order.id.value, 1, 'USD', now)]);
    expect(order.pullEvents()).toEqual([]);
    expect(order.status).toBe(OrderStatus.Placed);
    expect(() => order.addItem('Pencil', 1, usd(100))).toThrow(DomainError);
    expect(() => order.place(now)).toThrow('already been placed');
  });
});

describe('Money', () => {
  it('rejects negative values and invalid currencies', () => {
    expect(() => Money.fromCents(-1, 'USD')).toThrow(DomainError);
    expect(() => Money.fromCents(100, 'EURO')).toThrow(DomainError);
  });

  it('defaults to USD and converts amounts to cents', () => {
    const price = Money.fromAmount(0.29);
    expect(price.cents).toBe(29);
    expect(price.currency).toBe('USD');
  });

  it('refuses to add different currencies', () => {
    expect(() => usd(100).add(Money.fromCents(100, 'EUR'))).toThrow(DomainError);
  });
});

describe('OrderId', () => {
  it('accepts only UUIDs', () => {
    expect(() => OrderId.parse('42')).toThrow(DomainError);
    expect(OrderId.parse('6F9619FF-8B86-4D11-B42D-00C04FC964FF').value).toBe('6f9619ff-8b86-4d11-b42d-00c04fc964ff');
  });
});
