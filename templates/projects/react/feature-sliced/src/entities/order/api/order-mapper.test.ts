import { describe, expect, it } from 'vitest';
import { money } from '../model/money';
import { orderOf } from './order-mapper';

describe('orderOf', () => {
  it('maps the API contract to the model, with the order currency on every amount', () => {
    const order = orderOf({
      id: 'o-1',
      customerName: 'Ana',
      status: 'Placed',
      createdAt: '2026-01-02T10:00:00Z',
      total: 7,
      currency: 'USD',
      items: [{ productName: 'Caneta', quantity: 2, unitPrice: 3.5, subtotal: 7 }],
    });

    expect(order.status).toBe('Placed');
    expect(order.createdAt.toISOString()).toBe('2026-01-02T10:00:00.000Z');
    expect(order.total).toEqual(money(7, 'USD'));
    expect(order.items[0]?.unitPrice).toEqual(money(3.5, 'USD'));
  });
});
