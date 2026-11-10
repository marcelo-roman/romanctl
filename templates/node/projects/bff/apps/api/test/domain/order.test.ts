import { describe, expect, it } from 'vitest';
import { CatalogProduct } from '../../src/domain/catalog-product.js';
import { DomainError } from '../../src/domain/domain-error.js';
import { Money } from '../../src/domain/money.js';
import { Order } from '../../src/domain/order.js';
import { OrderStatus } from '../../src/domain/order-status.js';

const now = new Date('2026-01-01T12:00:00Z');
const brl = (cents: number) => Money.fromCents(cents, 'brl');

describe('Order', () => {
  it('requires a customer name', () => {
    expect(() => Order.create('  ', now)).toThrow(DomainError);
  });

  it('starts as a draft with the trimmed customer name and nothing to place', () => {
    const order = Order.create('  Ana ', now);
    expect(order.customerName).toBe('Ana');
    expect(order.status).toBe(OrderStatus.Draft);
    expect(order.total).toEqual(Money.fromCents(0, 'BRL'));
    expect(order.canAddItems).toBe(true);
    expect(order.canPlace).toBe(false);
  });

  it('sums the items into the total', () => {
    const order = Order.create('Ana', now);
    order.addItem('Caneta', 2, brl(350));
    order.addItem('Caderno', 1, brl(1200));
    expect(order.total.cents).toBe(1900);
    expect(order.currency).toBe('BRL');
    expect(order.canPlace).toBe(true);
  });

  it('requires a product name and a positive whole quantity', () => {
    const order = Order.create('Ana', now);
    expect(() => order.addItem(' ', 1, brl(100))).toThrow('nome do produto');
    expect(() => order.addItem('Caneta', 0, brl(100))).toThrow('maior que zero');
    expect(() => order.addItem('Caneta', 1.5, brl(100))).toThrow('número inteiro');
  });

  it('keeps every item in the same currency', () => {
    const order = Order.create('Ana', now);
    order.addItem('Caneta', 1, brl(100));
    expect(() => order.addItem('Pen', 1, Money.fromCents(100, 'USD'))).toThrow('mesma moeda');
  });

  it('places only a draft with items and then refuses changes', () => {
    const order = Order.create('Ana', now);
    expect(() => order.place()).toThrow('sem itens');
    order.addItem('Caneta', 1, brl(100));
    order.place();

    expect(order.status).toBe(OrderStatus.Placed);
    expect(order.canAddItems).toBe(false);
    expect(order.canPlace).toBe(false);
    expect(() => order.addItem('Lápis', 1, brl(100))).toThrow('não pode ser alterado');
    expect(() => order.place()).toThrow('já foi realizado');
  });
});

describe('Money', () => {
  it('rejects negative values and invalid currencies', () => {
    expect(() => Money.fromCents(-1, 'BRL')).toThrow(DomainError);
    expect(() => Money.fromCents(100, 'REAL')).toThrow(DomainError);
    expect(() => Money.fromCents(100, '12A')).toThrow(DomainError);
  });

  it('defaults to BRL and converts amounts to cents', () => {
    const price = Money.fromAmount(0.29);
    expect(price.cents).toBe(29);
    expect(price.currency).toBe('BRL');
  });
});

describe('CatalogProduct', () => {
  it('requires a name', () => {
    expect(() => CatalogProduct.create(' ', brl(100), now)).toThrow(DomainError);
    expect(CatalogProduct.create(' Caneta ', brl(350), now).name).toBe('Caneta');
  });
});
