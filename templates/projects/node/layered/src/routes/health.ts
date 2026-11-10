import type { FastifyPluginAsync } from 'fastify';

const healthSchema = {
  response: {
    200: { type: 'object', required: ['status'], properties: { status: { type: 'string' } } },
  },
} as const;

export const healthRoutes: FastifyPluginAsync = async (app) => {
  app.get('/health', { schema: healthSchema }, async () => ({ status: 'ok' }));
};
