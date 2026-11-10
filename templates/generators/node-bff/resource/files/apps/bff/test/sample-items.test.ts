import type { FastifyInstance } from 'fastify';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { buildApp } from '../src/app.js';
import { type FakeApi, problem, startFakeApi } from './fake-api.js';

describe('BFF sample items', () => {
  let api: FakeApi;
  let app: FastifyInstance;
  let stored: { id: string; name: string; createdAt: string }[];

  beforeEach(async () => {
    stored = [{ id: 'r1', name: 'Primeiro', createdAt: '2026-03-10T12:00:00.000Z' }];
    api = await startFakeApi((fake) => {
      fake.get('/sample-items', async () => stored);
      fake.post('/sample-items', async (request, reply) => {
        const { name } = request.body as { name: string };
        if (name.trim() === '') {
          return reply.status(422).send(problem(422, 'O nome é obrigatório.'));
        }
        const created = { id: 'r2', name, createdAt: '2026-03-11T12:00:00.000Z' };
        stored = [created, ...stored];
        return reply.status(201).send(created);
      });
      fake.delete('/sample-items/:id', async (request, reply) => {
        const { id } = request.params as { id: string };
        if (!stored.some((item) => item.id === id)) {
          return reply.status(404).send(problem(404, `Registro ${id} não encontrado.`));
        }
        stored = stored.filter((item) => item.id !== id);
        return reply.status(204).send();
      });
    });
    app = buildApp({ apiUrl: api.url });
    await app.ready();
  });

  afterEach(async () => {
    await app.close();
    await api.close();
  });

  it('shapes the list for the screen', async () => {
    const response = await app.inject({ method: 'GET', url: '/screens/sample-items' });
    expect(response.json()).toEqual({ sampleItems: [{ id: 'r1', name: 'Primeiro', createdOn: '10/03/2026' }] });
  });

  it('creates and deletes, answering the refreshed screen', async () => {
    const created = await app.inject({ method: 'POST', url: '/actions/sample-items', payload: { name: 'Segundo' } });
    expect(created.json().sampleItems.map((item: { name: string }) => item.name)).toEqual(['Segundo', 'Primeiro']);

    const deleted = await app.inject({ method: 'DELETE', url: '/actions/sample-items/r1' });
    expect(deleted.json().sampleItems.map((item: { id: string }) => item.id)).toEqual(['r2']);
  });

  it('translates API errors into messages', async () => {
    const empty = await app.inject({ method: 'POST', url: '/actions/sample-items', payload: { name: ' ' } });
    expect(empty.statusCode).toBe(422);
    expect(empty.json()).toEqual({ message: 'O nome é obrigatório.' });

    const missing = await app.inject({ method: 'DELETE', url: '/actions/sample-items/nada' });
    expect(missing.statusCode).toBe(404);
    expect(missing.json()).toEqual({ message: 'Registro não encontrado.' });
  });
});
