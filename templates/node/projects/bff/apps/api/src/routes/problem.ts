import type { FastifyError, FastifyReply, FastifyRequest } from 'fastify';
import { DomainError } from '../domain/domain-error.js';
import { NotFoundError } from '../services/errors.js';
import { validationMessage } from './validation-message.js';

export interface Problem {
  status: number;
  title: string;
  detail?: string;
}

const titles: Record<number, string> = {
  400: 'Bad request',
  404: 'Not found',
  415: 'Unsupported media type',
  422: 'Business rule violated',
  500: 'Internal error',
};

const requestErrors: Record<string, string> = {
  FST_ERR_CTP_INVALID_JSON_BODY: 'The body is not valid JSON.',
  FST_ERR_CTP_EMPTY_JSON_BODY: 'The body is empty.',
  FST_ERR_CTP_INVALID_MEDIA_TYPE: 'Send the body as application/json.',
};

function problem(status: number, detail: string): Problem {
  return { status, title: titles[status] ?? 'Error', detail };
}

function problemOf(error: FastifyError): Problem {
  if (error.validation) {
    return problem(400, validationMessage(error.validation, error.validationContext));
  }
  if (error instanceof NotFoundError) {
    return problem(404, error.message);
  }
  if (error instanceof DomainError) {
    return problem(422, error.message);
  }
  if (error.statusCode && error.statusCode >= 400 && error.statusCode < 500) {
    return problem(error.statusCode, requestErrors[error.code] ?? error.message);
  }
  return { status: 500, title: titles[500] ?? 'Error' };
}

export function sendProblem(error: FastifyError, request: FastifyRequest, reply: FastifyReply): FastifyReply {
  const result = problemOf(error);
  if (result.status >= 500) {
    request.log.error(error);
  }
  return reply.status(result.status).type('application/problem+json').send(result);
}

export function sendNotFound(request: FastifyRequest, reply: FastifyReply): FastifyReply {
  return reply
    .status(404)
    .type('application/problem+json')
    .send(problem(404, `The route ${request.method} ${request.url.split('?')[0]} does not exist.`));
}
