import fastifyStatic from '@fastify/static';
import type { FastifyInstance, FastifyReply, FastifyRequest } from 'fastify';
import { sendNotFound } from './problem.js';

const apiPrefixes = ['/screens/', '/actions/', '/health'];

function wantsPage(request: FastifyRequest): boolean {
  const path = request.url.split('?')[0] ?? '';
  return request.method === 'GET' && !apiPrefixes.some((prefix) => path.startsWith(prefix)) && !/\.[a-z0-9]+$/i.test(path);
}

export function serveWeb(app: FastifyInstance, root: string): void {
  app.register(fastifyStatic, { root, wildcard: false });
  app.setNotFoundHandler((request: FastifyRequest, reply: FastifyReply) => {
    if (wantsPage(request)) {
      return reply.sendFile('index.html');
    }
    return sendNotFound(request, reply);
  });
}
