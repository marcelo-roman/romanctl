import { describe, expect, it } from 'vitest';
import { money } from './money';
import { isDraft, itemCount, statusLabel, type Order } from './order';

const order: Order = {
  id: 'o-1',
  customerName: 'Ana',
  status: 'Draft',
  createdAt: new Date('2026-01-02T10:00:00Z'),
  total: money(12.5),
  items: [
    { productName: 'Pen', quantity: 2, unitPrice: money(3.5), subtotal: money(7) },
    { productName: 'Notebook', quantity: 1, unitPrice: money(5.5), subtotal: money(5.5) },
  ],
};

describe('order', () => {
  it('counts every unit of every item', () => {
    expect(itemCount(order)).toBe(3);
    expect(itemCount({ ...order, items: [] })).toBe(0);
  });

  it('only accepts changes while it is a draft', () => {
    expect(isDraft(order)).toBe(true);
    expect(isDraft({ ...order, status: 'Placed' })).toBe(false);
  });

  it('labels the status for display', () => {
    expect(statusLabel('Draft')).toBe('Draft');
    expect(statusLabel('Placed')).toBe('Placed');
  });
});
