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

    const added = await app.inject({ method: 'POST', url: `/orders/${id}/items`, payload: { productName: 'Caneta', quantity: 2, unitPrice: 3.5 } });
    expect(added.statusCode).toBe(200);
    expect(added.json()).toMatchObject({ total: 7, currency: 'BRL', canPlace: true });

    const placed = await app.inject({ method: 'POST', url: `/orders/${id}/place` });
    expect(placed.json()).toMatchObject({ status: 'Placed', canAddItems: false, canPlace: false });

    const read = await app.inject({ method: 'GET', url: `/orders/${id}` });
    expect(read.json().items).toEqual([{ id: expect.any(String), productName: 'Caneta', quantity: 2, unitPrice: 3.5, subtotal: 7 }]);

    const list = await app.inject({ method: 'GET', url: '/orders?limit=10' });
    expect(list.json().map((order: { id: string }) => order.id)).toEqual([id]);
  });

  it('turns business rules into 422 problems', async () => {
    const id = await createOrder();
    const cases: [InjectOptions, string][] = [
      [{ method: 'POST', url: '/orders', payload: { customerName: '  ' } }, 'O nome do cliente é obrigatório.'],
      [{ method: 'POST', url: `/orders/${id}/place` }, 'Um pedido sem itens não pode ser realizado.'],
      [{ method: 'POST', url: `/orders/${id}/items`, payload: { productName: 'Caneta', quantity: 0, unitPrice: 1 } }, 'A quantidade deve ser maior que zero.'],
      [{ method: 'POST', url: `/orders/${id}/items`, payload: { productName: 'Caneta', quantity: 1, unitPrice: -1 } }, 'O valor não pode ser negativo.'],
    ];
    for (const [request, detail] of cases) {
      const response = await app.inject(request);
      expect(response.statusCode, detail).toBe(422);
      expect(response.headers['content-type']).toContain('application/problem+json');
      expect(response.json()).toEqual({ status: 422, title: 'Regra de negócio violada', detail });
    }
  });

  it('refuses a second currency', async () => {
    const id = await createOrder();
    await app.inject({ method: 'POST', url: `/orders/${id}/items`, payload: { productName: 'Caneta', quantity: 1, unitPrice: 1 } });
    const response = await app.inject({ method: 'POST', url: `/orders/${id}/items`, payload: { productName: 'Pen', quantity: 1, unitPrice: 1, currency: 'USD' } });
    expect(response.statusCode).toBe(422);
    expect(response.json().detail).toBe('Todos os itens do pedido devem usar a mesma moeda.');
  });

  it('answers 400 for invalid requests', async () => {
    const cases: [InjectOptions, string][] = [
      [{ method: 'POST', url: '/orders', payload: {} }, 'O campo customerName é obrigatório.'],
      [{ method: 'POST', url: '/orders', payload: { customerName: { nome: 'Ana' } } }, 'O campo customerName deve ser um texto.'],
      [{ method: 'POST', url: '/orders', headers: { 'content-type': 'application/json' }, payload: '{' }, 'O corpo não é um JSON válido.'],
      [{ method: 'POST', url: `/orders/${missingId}/items`, payload: { productName: 'Caneta', quantity: 'dois', unitPrice: 1 } }, 'O campo quantity deve ser um número.'],
      [{ method: 'GET', url: '/orders/123' }, 'O campo id deve estar no formato UUID.'],
    ];
    for (const [request, detail] of cases) {
      const response = await app.inject(request);
      expect(response.statusCode, detail).toBe(400);
      expect(response.json()).toEqual({ status: 400, title: 'Requisição inválida', detail });
    }
  });

  it('answers 404 for unknown orders and routes', async () => {
    const missing = await app.inject({ method: 'GET', url: `/orders/${missingId}` });
    expect(missing.statusCode).toBe(404);
    expect(missing.json()).toEqual({ status: 404, title: 'Não encontrado', detail: `Pedido ${missingId} não encontrado.` });
    expect((await app.inject({ method: 'GET', url: '/nada' })).statusCode).toBe(404);
  });
});
