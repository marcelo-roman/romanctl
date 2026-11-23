import { describe, expect, it } from 'vitest';
import { CatalogProduct } from '../../src/domain/catalog-product.js';
import { DomainError } from '../../src/domain/domain-error.js';
import { Money } from '../../src/domain/money.js';
import { Order } from '../../src/domain/order.js';
import { OrderStatus } from '../../src/domain/order-status.js';

const now = new Date('2026-01-01T12:00:00Z');
const usd = (cents: number) => Money.fromCents(cents, 'usd');

describe('Order', () => {
  it('requires a customer name', () => {
    expect(() => Order.create('  ', now)).toThrow(DomainError);
  });

  it('starts as a draft with the trimmed customer name and nothing to place', () => {
    const order = Order.create('  Ana ', now);
    expect(order.customerName).toBe('Ana');
    expect(order.status).toBe(OrderStatus.Draft);
    expect(order.total).toEqual(Money.fromCents(0, 'USD'));
    expect(order.canAddItems).toBe(true);
    expect(order.canPlace).toBe(false);
  });

  it('sums the items into the total', () => {
    const order = Order.create('Ana', now);
    order.addItem('Pen', 2, usd(350));
    order.addItem('Notebook', 1, usd(1200));
    expect(order.total.cents).toBe(1900);
    expect(order.currency).toBe('USD');
    expect(order.canPlace).toBe(true);
  });

  it('requires a product name and a positive whole quantity', () => {
    const order = Order.create('Ana', now);
    expect(() => order.addItem(' ', 1, usd(100))).toThrow('product name');
    expect(() => order.addItem('Pen', 0, usd(100))).toThrow('greater than zero');
    expect(() => order.addItem('Pen', 1.5, usd(100))).toThrow('whole number');
  });

  it('keeps every item in the same currency', () => {
    const order = Order.create('Ana', now);
    order.addItem('Pen', 1, usd(100));
    expect(() => order.addItem('Pen', 1, Money.fromCents(100, 'EUR'))).toThrow('same currency');
  });

  it('places only a draft with items and then refuses changes', () => {
    const order = Order.create('Ana', now);
    expect(() => order.place()).toThrow('without items');
    order.addItem('Pen', 1, usd(100));
    order.place();

    expect(order.status).toBe(OrderStatus.Placed);
    expect(order.canAddItems).toBe(false);
    expect(order.canPlace).toBe(false);
    expect(() => order.addItem('Pencil', 1, usd(100))).toThrow('cannot be changed');
    expect(() => order.place()).toThrow('already placed');
  });
});

describe('Money', () => {
  it('rejects negative values and invalid currencies', () => {
    expect(() => Money.fromCents(-1, 'USD')).toThrow(DomainError);
    expect(() => Money.fromCents(100, 'REAL')).toThrow(DomainError);
    expect(() => Money.fromCents(100, '12A')).toThrow(DomainError);
  });

  it('defaults to USD and converts amounts to cents', () => {
    const price = Money.fromAmount(0.29);
    expect(price.cents).toBe(29);
    expect(price.currency).toBe('USD');
  });
});

describe('CatalogProduct', () => {
  it('requires a name', () => {
    expect(() => CatalogProduct.create(' ', usd(100), now)).toThrow(DomainError);
    expect(CatalogProduct.create(' Pen ', usd(350), now).name).toBe('Pen');
  });
});
