import type { FastifyError, FastifyReply, FastifyRequest } from 'fastify';
import { uiError } from './ui-error.js';

export function sendError(error: FastifyError, request: FastifyRequest, reply: FastifyReply): FastifyReply {
  const { status, message } = uiError(error);
  if (status >= 500) {
    request.log.error(error);
  }
  return reply.status(status).send({ message });
}

export function sendNotFound(request: FastifyRequest, reply: FastifyReply): FastifyReply {
  return reply.status(404).send({ message: `A rota ${request.method} ${request.url.split('?')[0]} não existe.` });
}
