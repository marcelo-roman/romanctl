import type { FastifyInstance } from 'fastify';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { buildApp } from '../src/app.js';
import { type FakeApi, problem, startFakeApi } from './fake-api.js';

const orderId = '6f9619ff-8b86-4d11-b42d-00c04fc964ff';
const missingId = '00000000-0000-4000-8000-000000000000';

const caneta = { id: 'p-caneta', name: 'Caneta', price: 3.5, currency: 'BRL' };
const caderno = { id: 'p-caderno', name: 'Caderno', price: 12, currency: 'BRL' };

function apiOrder(overrides: Record<string, unknown> = {}) {
  return {
    id: orderId,
    customerName: 'Ana',
    status: 'Draft',
    createdAt: '2026-03-10T12:00:00.000Z',
    currency: 'BRL',
    total: 19,
    canAddItems: true,
    canPlace: true,
    items: [
      { id: 'i1', productName: 'Caneta', quantity: 2, unitPrice: 3.5, subtotal: 7 },
      { id: 'i2', productName: 'Caderno', quantity: 1, unitPrice: 12, subtotal: 12 },
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
        return id === orderId ? apiOrder() : reply.status(404).send(problem(404, `Pedido ${id} não encontrado.`));
      });
      fake.post('/orders', async (_request, reply) => reply.status(201).send(apiOrder({ items: [], total: 0, canPlace: false })));
      fake.post('/orders/:id/items', async (request, reply) => {
        const body = request.body as { quantity: number };
        if (body.quantity <= 0) {
          return reply.status(422).send(problem(422, 'A quantidade deve ser maior que zero.'));
        }
        return apiOrder();
      });
      fake.post('/orders/:id/place', async () => apiOrder({ status: 'Placed', canAddItems: false, canPlace: false }));
      fake.get('/catalog/products', async () => [caderno, caneta]);
      fake.get('/catalog/products/:id', async (request, reply) => {
        const { id } = request.params as { id: string };
        const product = [caneta, caderno].find((p) => p.id === id);
        return product ?? reply.status(404).send(problem(404, `Produto ${id} não encontrado no catálogo.`));
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
        { id: orderId, customerName: 'Ana', status: 'Rascunho', itemCount: 3, total: 'R$ 19,00' },
        { id: 'o2', customerName: 'Bia', status: 'Realizado', itemCount: 0, total: 'R$ 0,00' },
      ],
    });
  });

  it('joins the order with product suggestions from the catalog', async () => {
    const response = await app.inject({ method: 'GET', url: `/screens/orders/${orderId}` });

    expect(response.json()).toEqual({
      order: {
        id: orderId,
        customerName: 'Ana',
        status: 'Rascunho',
        total: 'R$ 19,00',
        canAddItems: true,
        canPlace: true,
        items: [
          { id: 'i1', productName: 'Caneta', quantity: 2, unitPrice: 'R$ 3,50', subtotal: 'R$ 7,00' },
          { id: 'i2', productName: 'Caderno', quantity: 1, unitPrice: 'R$ 12,00', subtotal: 'R$ 12,00' },
        ],
      },
      suggestions: [
        { id: 'p-caderno', name: 'Caderno', price: 'R$ 12,00' },
        { id: 'p-caneta', name: 'Caneta', price: 'R$ 3,50' },
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
    const response = await app.inject({ method: 'POST', url: `/actions/orders/${orderId}/items`, payload: { productId: 'p-caneta', quantity: 2 } });

    expect(response.statusCode).toBe(200);
    expect(response.json().order.items).toHaveLength(2);
    expect(api.calls.find((call) => call.url === `/orders/${orderId}/items`)?.body).toEqual({
      productName: 'Caneta',
      quantity: 2,
      unitPrice: 3.5,
      currency: 'BRL',
    });
  });

  it('places the order and hides suggestions once it cannot change', async () => {
    const response = await app.inject({ method: 'POST', url: `/actions/orders/${orderId}/place` });

    expect(response.json().order).toMatchObject({ status: 'Realizado', canAddItems: false, canPlace: false });
    expect(response.json().suggestions).toEqual([]);
  });

  it('translates API errors into messages for the screen', async () => {
    const rule = await app.inject({ method: 'POST', url: `/actions/orders/${orderId}/items`, payload: { productId: 'p-caneta', quantity: 0 } });
    expect(rule.statusCode).toBe(422);
    expect(rule.json()).toEqual({ message: 'A quantidade deve ser maior que zero.' });

    const order = await app.inject({ method: 'GET', url: `/screens/orders/${missingId}` });
    expect(order.statusCode).toBe(404);
    expect(order.json()).toEqual({ message: 'Pedido não encontrado.' });

    const product = await app.inject({ method: 'POST', url: `/actions/orders/${orderId}/items`, payload: { productId: 'nenhum', quantity: 1 } });
    expect(product.statusCode).toBe(404);
    expect(product.json()).toEqual({ message: 'Produto não encontrado no catálogo.' });
  });

  it('checks the shape of the request before calling the API', async () => {
    const response = await app.inject({ method: 'POST', url: `/actions/orders/${orderId}/items`, payload: { productId: 'p-caneta', quantity: 'dois' } });

    expect(response.statusCode).toBe(400);
    expect(response.json()).toEqual({ message: 'Confira os dados informados: o campo quantidade deve ser um número inteiro.' });
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
    expect(response.json()).toEqual({ message: 'O serviço de pedidos não respondeu. Tente de novo em instantes.' });
    await app.close();
  });

  it('answers the health check on its own', async () => {
    const app = buildApp({ apiUrl: 'http://127.0.0.1:9' });
    expect((await app.inject({ method: 'GET', url: '/health' })).json()).toEqual({ status: 'ok' });
    await app.close();
  });
});
