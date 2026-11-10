import type { FastifyInstance } from 'fastify';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { startApp } from './app.js';

describe('health and unknown routes', () => {
  let app: FastifyInstance;

  beforeEach(async () => {
    app = await startApp();
  });

  afterEach(() => app.close());

  it('answers the health check', async () => {
    const response = await app.inject({ method: 'GET', url: '/health' });

    expect(response.statusCode).toBe(200);
    expect(response.json()).toEqual({ status: 'ok' });
  });

  it('answers unknown routes with a 404 problem', async () => {
    const response = await app.inject({ method: 'GET', url: '/nada?x=1' });

    expect(response.statusCode).toBe(404);
    expect(response.headers['content-type']).toContain('application/problem+json');
    expect(response.json()).toEqual({ status: 404, title: 'Não encontrado', detail: 'A rota GET /nada não existe.' });
  });
});
