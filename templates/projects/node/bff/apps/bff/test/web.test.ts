import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import type { FastifyInstance } from 'fastify';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { buildApp } from '../src/app.js';

describe('BFF serving the built web', () => {
  let dist: string;
  let app: FastifyInstance;

  beforeEach(async () => {
    dist = mkdtempSync(join(tmpdir(), 'web-dist-'));
    writeFileSync(join(dist, 'index.html'), '<div id="root"></div>');
    writeFileSync(join(dist, 'app.js'), 'console.log(1)');
    app = buildApp({ apiUrl: 'http://127.0.0.1:9', webDist: dist });
    await app.ready();
  });

  afterEach(async () => {
    await app.close();
    rmSync(dist, { recursive: true, force: true });
  });

  it('serves files and falls back to index.html for pages', async () => {
    expect((await app.inject({ method: 'GET', url: '/app.js' })).body).toBe('console.log(1)');

    const page = await app.inject({ method: 'GET', url: '/orders/123' });
    expect(page.statusCode).toBe(200);
    expect(page.body).toContain('id="root"');
  });

  it('keeps unknown BFF routes and assets as 404', async () => {
    expect((await app.inject({ method: 'GET', url: '/screens/nada' })).statusCode).toBe(404);
    expect((await app.inject({ method: 'GET', url: '/missing.js' })).statusCode).toBe(404);
  });
});
