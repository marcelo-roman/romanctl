import type { INestApplication } from '@nestjs/common';
import request from 'supertest';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { startApp } from './app.js';

describe('orders API', () => {
  let app: INestApplication;
  let http: ReturnType<typeof request>;

  beforeEach(async () => {
    app = await startApp();
    http = request(app.getHttpServer());
  });

  afterEach(() => app.close());

  it('answers the health check', async () => {
    await http.get('/health').expect(200, { status: 'ok' });
  });

  it('runs an order from draft to placed', async () => {
    const created = await http.post('/orders').send({ customerName: 'Ana' }).expect(201);
    const id: string = created.body.id;
    expect(created.headers.location).toBe(`/orders/${id}`);

    await http.post(`/orders/${id}/items`).send({ productName: 'Caneta', quantity: 2, unitPrice: 3.5 }).expect(204);
    await http.post(`/orders/${id}/place`).expect(204);

    const order = await http.get(`/orders/${id}`).expect(200);
    expect(order.body).toMatchObject({ id, customerName: 'Ana', status: 'Placed', total: 7, currency: 'BRL' });
    expect(order.body.items).toEqual([{ productName: 'Caneta', quantity: 2, unitPrice: 3.5, subtotal: 7 }]);

    const list = await http.get('/orders?limit=10').expect(200);
    expect(list.body.map((o: { id: string }) => o.id)).toEqual([id]);
  });

  it('turns business rules into 422 problems', async () => {
    const empty = await http.post('/orders').send({ customerName: '' }).expect(422);
    expect(empty.headers['content-type']).toContain('application/problem+json');
    expect(empty.body).toEqual({ status: 422, title: 'Regra de negócio violada', detail: 'O nome do cliente é obrigatório.' });

    const { body } = await http.post('/orders').send({ customerName: 'Ana' }).expect(201);
    await http.post(`/orders/${body.id}/place`).expect(422);
    await http.post(`/orders/${body.id}/items`).send({ productName: 'Caneta', quantity: 0, unitPrice: 1 }).expect(422);
  });

  it('answers 404 for unknown orders', async () => {
    const missing = await http.get('/orders/00000000-0000-4000-8000-000000000000').expect(404);
    expect(missing.body.title).toBe('Não encontrado');
  });

  it('answers 400 for invalid bodies', async () => {
    const malformed = await http.post('/orders').set('Content-Type', 'application/json').send('{').expect(400);
    expect(malformed.body).toMatchObject({ status: 400, title: 'Requisição inválida' });
    await http.post('/orders').send({ customerName: 42 }).expect(400);
    const { body } = await http.post('/orders').send({ customerName: 'Ana' }).expect(201);
    const bad = await http.post(`/orders/${body.id}/items`).send({ productName: 'Caneta', quantity: 'dois', unitPrice: 1 }).expect(400);
    expect(bad.body).toEqual({ status: 400, title: 'Requisição inválida', detail: 'O campo quantity deve ser um número.' });
  });
});
