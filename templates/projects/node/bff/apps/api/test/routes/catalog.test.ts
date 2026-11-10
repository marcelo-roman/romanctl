import type { FastifyInstance } from 'fastify';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { startApp } from './app.js';

describe('catalog API', () => {
  let app: FastifyInstance;

  beforeEach(async () => {
    app = await startApp();
  });

  afterEach(() => app.close());

  it('adds, lists and reads products', async () => {
    const created = await app.inject({ method: 'POST', url: '/catalog/products', payload: { name: 'Caneta', price: 3.5 } });
    expect(created.statusCode).toBe(201);
    const product = created.json();
    expect(product).toEqual({ id: product.id, name: 'Caneta', price: 3.5, currency: 'BRL' });
    expect(created.headers.location).toBe(`/catalog/products/${product.id}`);

    await app.inject({ method: 'POST', url: '/catalog/products', payload: { name: 'Caderno', price: 12 } });
    const list = await app.inject({ method: 'GET', url: '/catalog/products' });
    expect(list.json().map((p: { name: string }) => p.name)).toEqual(['Caderno', 'Caneta']);

    const read = await app.inject({ method: 'GET', url: `/catalog/products/${product.id}` });
    expect(read.json()).toEqual(product);
  });

  it('answers problems for invalid and unknown products', async () => {
    const invalid = await app.inject({ method: 'POST', url: '/catalog/products', payload: { name: 'Caneta', price: 1, currency: 'REAL' } });
    expect(invalid.statusCode).toBe(422);

    const missing = await app.inject({ method: 'GET', url: '/catalog/products/00000000-0000-4000-8000-000000000000' });
    expect(missing.statusCode).toBe(404);
    expect(missing.json().detail).toContain('não encontrado no catálogo');
  });
});
