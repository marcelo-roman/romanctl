import type { FastifyInstance } from 'fastify';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { buildApp } from '../src/app.js';
import { type FakeApi, problem, startFakeApi } from './fake-api.js';

const orderId = '6f9619ff-8b86-4d11-b42d-00c04fc964ff';
const missingId = '00000000-0000-4000-8000-000000000000';

const pen = { id: 'p-pen', name: 'Pen', price: 3.5, currency: 'USD' };
const notebook = { id: 'p-notebook', name: 'Notebook', price: 12, currency: 'USD' };

function apiOrder(overrides: Record<string, unknown> = {}) {
  return {
    id: orderId,
    customerName: 'Ana',
    status: 'Draft',
    createdAt: '2026-03-10T12:00:00.000Z',
    currency: 'USD',
    total: 19,
    canAddItems: true,
    canPlace: true,
    items: [
      { id: 'i1', productName: 'Pen', quantity: 2, unitPrice: 3.5, subtotal: 7 },
      { id: 'i2', productName: 'Notebook', quantity: 1, unitPrice: 12, subtotal: 12 },
    ],
    ...overrides,
  };
}

describe('BFF orders', () => {
  let api: FakeApi;
  let app: FastifyInstance;

  beforeEach(async () => {
    api = await startFakeApi((fake) => {
      fake.get('/orders', async () => [apiOrder(), apiOrder({ id: 'o2', customerName: 'Bia', status: 'Placed', total: 0, items: [] })]);
      fake.get('/orders/:id', async (request, reply) => {
        const { id } = request.params as { id: string };
        return id === orderId ? apiOrder() : reply.status(404).send(problem(404, `Order ${id} not found.`));
      });
      fake.post('/orders', async (_request, reply) => reply.status(201).send(apiOrder({ items: [], total: 0, canPlace: false })));
      fake.post('/orders/:id/items', async (request, reply) => {
        const body = request.body as { quantity: number };
        if (body.quantity <= 0) {
          return reply.status(422).send(problem(422, 'The quantity must be greater than zero.'));
        }
        return apiOrder();
      });
      fake.post('/orders/:id/place', async () => apiOrder({ status: 'Placed', canAddItems: false, canPlace: false }));
      fake.get('/catalog/products', async () => [notebook, pen]);
      fake.get('/catalog/products/:id', async (request, reply) => {
        const { id } = request.params as { id: string };
        const product = [pen, notebook].find((p) => p.id === id);
        return product ?? reply.status(404).send(problem(404, `Product ${id} not found in the catalog.`));
      });
    });
    app = buildApp({ apiUrl: api.url });
    await app.ready();
  });

  afterEach(async () => {
    await app.close();
    await api.close();
  });

  it('shapes the orders list for the screen', async () => {
    const response = await app.inject({ method: 'GET', url: '/screens/orders' });

    expect(response.statusCode).toBe(200);
    expect(response.json()).toEqual({
      orders: [
        { id: orderId, customerName: 'Ana', status: 'Draft', itemCount: 3, total: '$19.00' },
        { id: 'o2', customerName: 'Bia', status: 'Placed', itemCount: 0, total: '$0.00' },
      ],
    });
  });

  it('joins the order with product suggestions from the catalog', async () => {
    const response = await app.inject({ method: 'GET', url: `/screens/orders/${orderId}` });

    expect(response.json()).toEqual({
      order: {
        id: orderId,
        customerName: 'Ana',
        status: 'Draft',
        total: '$19.00',
        canAddItems: true,
        canPlace: true,
        items: [
          { id: 'i1', productName: 'Pen', quantity: 2, unitPrice: '$3.50', subtotal: '$7.00' },
          { id: 'i2', productName: 'Notebook', quantity: 1, unitPrice: '$12.00', subtotal: '$12.00' },
        ],
      },
      suggestions: [
        { id: 'p-notebook', name: 'Notebook', price: '$12.00' },
        { id: 'p-pen', name: 'Pen', price: '$3.50' },
      ],
    });
  });

  it('creates an order and answers only its id', async () => {
    const response = await app.inject({ method: 'POST', url: '/actions/orders', payload: { customerName: 'Ana' } });

    expect(response.statusCode).toBe(201);
    expect(response.json()).toEqual({ id: orderId });
    expect(api.calls.at(-1)).toMatchObject({ method: 'POST', url: '/orders', body: { customerName: 'Ana' } });
  });

  it('resolves the product name and price from the catalog before adding the item', async () => {
    const response = await app.inject({ method: 'POST', url: `/actions/orders/${orderId}/items`, payload: { productId: 'p-pen', quantity: 2 } });

    expect(response.statusCode).toBe(200);
    expect(response.json().order.items).toHaveLength(2);
    expect(api.calls.find((call) => call.url === `/orders/${orderId}/items`)?.body).toEqual({
      productName: 'Pen',
      quantity: 2,
      unitPrice: 3.5,
      currency: 'USD',
    });
  });

  it('places the order and hides suggestions once it cannot change', async () => {
    const response = await app.inject({ method: 'POST', url: `/actions/orders/${orderId}/place` });

    expect(response.json().order).toMatchObject({ status: 'Placed', canAddItems: false, canPlace: false });
    expect(response.json().suggestions).toEqual([]);
  });

  it('translates API errors into messages for the screen', async () => {
    const rule = await app.inject({ method: 'POST', url: `/actions/orders/${orderId}/items`, payload: { productId: 'p-pen', quantity: 0 } });
    expect(rule.statusCode).toBe(422);
    expect(rule.json()).toEqual({ message: 'The quantity must be greater than zero.' });

    const order = await app.inject({ method: 'GET', url: `/screens/orders/${missingId}` });
    expect(order.statusCode).toBe(404);
    expect(order.json()).toEqual({ message: 'Order not found.' });

    const product = await app.inject({ method: 'POST', url: `/actions/orders/${orderId}/items`, payload: { productId: 'none', quantity: 1 } });
    expect(product.statusCode).toBe(404);
    expect(product.json()).toEqual({ message: 'Product not found in the catalog.' });
  });

  it('checks the shape of the request before calling the API', async () => {
    const response = await app.inject({ method: 'POST', url: `/actions/orders/${orderId}/items`, payload: { productId: 'p-pen', quantity: 'two' } });

    expect(response.statusCode).toBe(400);
    expect(response.json()).toEqual({ message: 'Check the data you entered: the quantity field must be a whole number.' });
    expect(api.calls).toEqual([]);
  });
});

describe('BFF without the API', () => {
  it('answers 502 with a friendly message', async () => {
    const api = await startFakeApi(() => undefined);
    const url = api.url;
    await api.close();
    const app = buildApp({ apiUrl: url });

    const response = await app.inject({ method: 'GET', url: '/screens/orders' });

    expect(response.statusCode).toBe(502);
    expect(response.json()).toEqual({ message: 'The orders service did not answer. Try again in a moment.' });
    await app.close();
  });

  it('answers the health check on its own', async () => {
    const app = buildApp({ apiUrl: 'http://127.0.0.1:9' });
    expect((await app.inject({ method: 'GET', url: '/health' })).json()).toEqual({ status: 'ok' });
    await app.close();
  });
});
