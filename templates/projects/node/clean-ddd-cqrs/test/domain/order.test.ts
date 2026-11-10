import { describe, expect, it } from 'vitest';
import { Money } from '../../src/domain/orders/money.js';
import { Order } from '../../src/domain/orders/order.js';
import { OrderId } from '../../src/domain/orders/order-id.js';
import { OrderPlaced } from '../../src/domain/orders/order-placed.js';
import { OrderStatus } from '../../src/domain/orders/order-status.js';
import { DomainError } from '../../src/domain/shared/domain-error.js';

const now = new Date('2026-01-01T12:00:00Z');
const brl = (cents: number) => Money.fromCents(cents, 'brl');

describe('Order', () => {
  it('requires a customer name', () => {
    expect(() => Order.create('  ', now)).toThrow(DomainError);
  });

  it('starts as a draft with the trimmed customer name', () => {
    const order = Order.create('  Ana ', now);
    expect(order.customerName).toBe('Ana');
    expect(order.status).toBe(OrderStatus.Draft);
    expect(order.total).toEqual(Money.fromCents(0, 'BRL'));
  });

  it('sums the items into the total', () => {
    const order = Order.create('Ana', now);
    order.addItem('Caneta', 2, brl(350));
    order.addItem('Caderno', 1, brl(1200));
    expect(order.total.cents).toBe(1900);
    expect(order.total.currency).toBe('BRL');
  });

  it('requires a product name and a positive whole quantity', () => {
    const order = Order.create('Ana', now);
    expect(() => order.addItem(' ', 1, brl(100))).toThrow(DomainError);
    expect(() => order.addItem('Caneta', 0, brl(100))).toThrow(DomainError);
    expect(() => order.addItem('Caneta', 1.5, brl(100))).toThrow(DomainError);
  });

  it('keeps every item in the same currency', () => {
    const order = Order.create('Ana', now);
    order.addItem('Caneta', 1, brl(100));
    expect(() => order.addItem('Pen', 1, Money.fromCents(100, 'USD'))).toThrow('mesma moeda');
  });

  it('raises OrderPlaced once and then refuses changes', () => {
    const order = Order.create('Ana', now);
    expect(() => order.place(now)).toThrow('sem itens');
    order.addItem('Caneta', 1, brl(100));
    order.place(now);

    const events = order.pullEvents();
    expect(events).toEqual([new OrderPlaced(order.id.value, 1, 'BRL', now)]);
    expect(order.pullEvents()).toEqual([]);
    expect(order.status).toBe(OrderStatus.Placed);
    expect(() => order.addItem('Lápis', 1, brl(100))).toThrow(DomainError);
    expect(() => order.place(now)).toThrow('já foi realizado');
  });
});

describe('Money', () => {
  it('rejects negative values and invalid currencies', () => {
    expect(() => Money.fromCents(-1, 'BRL')).toThrow(DomainError);
    expect(() => Money.fromCents(100, 'REAL')).toThrow(DomainError);
  });

  it('defaults to BRL and converts amounts to cents', () => {
    const price = Money.fromAmount(0.29);
    expect(price.cents).toBe(29);
    expect(price.currency).toBe('BRL');
  });

  it('refuses to add different currencies', () => {
    expect(() => brl(100).add(Money.fromCents(100, 'USD'))).toThrow(DomainError);
  });
});

describe('OrderId', () => {
  it('accepts only UUIDs', () => {
    expect(() => OrderId.parse('42')).toThrow(DomainError);
    expect(OrderId.parse('6F9619FF-8B86-4D11-B42D-00C04FC964FF').value).toBe('6f9619ff-8b86-4d11-b42d-00c04fc964ff');
  });
});
