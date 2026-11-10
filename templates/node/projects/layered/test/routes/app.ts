import type { FastifyInstance } from 'fastify';
import { buildApp } from '../../src/app.js';
import { memoryRepositories } from '../../src/repositories/memory/memory-repositories.js';

export const now = new Date('2026-03-10T12:00:00Z');

export async function startApp(): Promise<FastifyInstance> {
  const app = buildApp(memoryRepositories(), { clock: () => now });
  await app.ready();
  return app;
}
