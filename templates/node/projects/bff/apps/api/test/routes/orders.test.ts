import type { FastifyInstance, InjectOptions } from 'fastify';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { now, startApp } from './app.js';

const missingId = '00000000-0000-4000-8000-000000000000';

describe('orders API', () => {
  let app: FastifyInstance;

  beforeEach(async () => {
    app = await startApp();
  });

  afterEach(() => app.close());

  async function createOrder(): Promise<string> {
    const response = await app.inject({ method: 'POST', url: '/orders', payload: { customerName: 'Ana' } });
    expect(response.statusCode).toBe(201);
    return response.json().id;
  }

  it('answers the health check', async () => {
    const response = await app.inject({ method: 'GET', url: '/health' });
    expect(response.json()).toEqual({ status: 'ok' });
  });

  it('runs an order from draft to placed', async () => {
    const created = await app.inject({ method: 'POST', url: '/orders', payload: { customerName: ' Ana ' } });
    expect(created.statusCode).toBe(201);
    const { id } = created.json();
    expect(created.headers.location).toBe(`/orders/${id}`);
    expect(created.json()).toMatchObject({ customerName: 'Ana', status: 'Draft', createdAt: now.toISOString(), canAddItems: true, canPlace: false });

    const added = await app.inject({ method: 'POST', url: `/orders/${id}/items`, payload: { productName: 'Pen', quantity: 2, unitPrice: 3.5 } });
    expect(added.statusCode).toBe(200);
    expect(added.json()).toMatchObject({ total: 7, currency: 'USD', canPlace: true });

    const placed = await app.inject({ method: 'POST', url: `/orders/${id}/place` });
    expect(placed.json()).toMatchObject({ status: 'Placed', canAddItems: false, canPlace: false });

    const read = await app.inject({ method: 'GET', url: `/orders/${id}` });
    expect(read.json().items).toEqual([{ id: expect.any(String), productName: 'Pen', quantity: 2, unitPrice: 3.5, subtotal: 7 }]);

    const list = await app.inject({ method: 'GET', url: '/orders?limit=10' });
    expect(list.json().map((order: { id: string }) => order.id)).toEqual([id]);
  });

  it('turns business rules into 422 problems', async () => {
    const id = await createOrder();
    const cases: [InjectOptions, string][] = [
      [{ method: 'POST', url: '/orders', payload: { customerName: '  ' } }, 'The customer name is required.'],
      [{ method: 'POST', url: `/orders/${id}/place` }, 'An order without items cannot be placed.'],
      [{ method: 'POST', url: `/orders/${id}/items`, payload: { productName: 'Pen', quantity: 0, unitPrice: 1 } }, 'The quantity must be greater than zero.'],
      [{ method: 'POST', url: `/orders/${id}/items`, payload: { productName: 'Pen', quantity: 1, unitPrice: -1 } }, 'The amount cannot be negative.'],
    ];
    for (const [request, detail] of cases) {
      const response = await app.inject(request);
      expect(response.statusCode, detail).toBe(422);
      expect(response.headers['content-type']).toContain('application/problem+json');
      expect(response.json()).toEqual({ status: 422, title: 'Business rule violated', detail });
    }
  });

  it('refuses a second currency', async () => {
    const id = await createOrder();
    await app.inject({ method: 'POST', url: `/orders/${id}/items`, payload: { productName: 'Pen', quantity: 1, unitPrice: 1 } });
    const response = await app.inject({ method: 'POST', url: `/orders/${id}/items`, payload: { productName: 'Pen', quantity: 1, unitPrice: 1, currency: 'EUR' } });
    expect(response.statusCode).toBe(422);
    expect(response.json().detail).toBe('All items of the order must use the same currency.');
  });

  it('answers 400 for invalid requests', async () => {
    const cases: [InjectOptions, string][] = [
      [{ method: 'POST', url: '/orders', payload: {} }, 'The field customerName is required.'],
      [{ method: 'POST', url: '/orders', payload: { customerName: { name: 'Ana' } } }, 'The field customerName must be a string.'],
      [{ method: 'POST', url: '/orders', headers: { 'content-type': 'application/json' }, payload: '{' }, 'The body is not valid JSON.'],
      [{ method: 'POST', url: `/orders/${missingId}/items`, payload: { productName: 'Pen', quantity: 'two', unitPrice: 1 } }, 'The field quantity must be a number.'],
      [{ method: 'GET', url: '/orders/123' }, 'The field id must be in the UUID format.'],
    ];
    for (const [request, detail] of cases) {
      const response = await app.inject(request);
      expect(response.statusCode, detail).toBe(400);
      expect(response.json()).toEqual({ status: 400, title: 'Bad request', detail });
    }
  });

  it('answers 404 for unknown orders and routes', async () => {
    const missing = await app.inject({ method: 'GET', url: `/orders/${missingId}` });
    expect(missing.statusCode).toBe(404);
    expect(missing.json()).toEqual({ status: 404, title: 'Not found', detail: `Order ${missingId} not found.` });
    expect((await app.inject({ method: 'GET', url: '/nothing' })).statusCode).toBe(404);
  });
});
