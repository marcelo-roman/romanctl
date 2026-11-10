import type { FastifyInstance } from 'fastify';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { startApp } from './app.js';

const missingId = '00000000-0000-4000-8000-000000000000';

describe('/sample-items', () => {
  let app: FastifyInstance;

  beforeEach(async () => {
    app = await startApp();
  });

  afterEach(() => app.close());

  it('creates, reads, renames, lists and deletes', async () => {
    const created = await app.inject({ method: 'POST', url: '/sample-items', payload: { name: 'Primeiro' } });
    expect(created.statusCode).toBe(201);
    const { id } = created.json();
    expect(created.headers.location).toBe(`/sample-items/${id}`);

    const renamed = await app.inject({ method: 'PUT', url: `/sample-items/${id}`, payload: { name: 'Segundo' } });
    expect(renamed.json()).toMatchObject({ id, name: 'Segundo' });

    const read = await app.inject({ method: 'GET', url: `/sample-items/${id}` });
    expect(read.json()).toEqual(renamed.json());

    const list = await app.inject({ method: 'GET', url: '/sample-items?limit=10' });
    expect(list.json().map((item: { id: string }) => item.id)).toEqual([id]);

    expect((await app.inject({ method: 'DELETE', url: `/sample-items/${id}` })).statusCode).toBe(204);
    expect((await app.inject({ method: 'GET', url: `/sample-items/${id}` })).statusCode).toBe(404);
  });

  it('answers problems for invalid and unknown records', async () => {
    const empty = await app.inject({ method: 'POST', url: '/sample-items', payload: { name: '   ' } });
    expect(empty.statusCode).toBe(422);
    expect(empty.json()).toEqual({ status: 422, title: 'Regra de negócio violada', detail: 'O nome é obrigatório.' });

    const missing = await app.inject({ method: 'GET', url: `/sample-items/${missingId}` });
    expect(missing.statusCode).toBe(404);
    expect(missing.headers['content-type']).toContain('application/problem+json');
  });
});
