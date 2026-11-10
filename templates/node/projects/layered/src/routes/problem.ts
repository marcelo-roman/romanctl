import type { FastifyError, FastifyReply, FastifyRequest } from 'fastify';
import { ConflictError, NotFoundError, RuleViolationError, ValidationError } from '../services/errors.js';
import { validationMessage } from './validation-message.js';

export interface Problem {
  status: number;
  title: string;
  detail?: string;
}

const titles: Record<number, string> = {
  400: 'Requisição inválida',
  404: 'Não encontrado',
  409: 'Conflito',
  415: 'Tipo de conteúdo não suportado',
  422: 'Regra de negócio violada',
  500: 'Erro interno',
};

const requestErrors: Record<string, string> = {
  FST_ERR_CTP_INVALID_JSON_BODY: 'O corpo não é um JSON válido.',
  FST_ERR_CTP_EMPTY_JSON_BODY: 'O corpo está vazio.',
  FST_ERR_CTP_INVALID_MEDIA_TYPE: 'Envie o corpo como application/json.',
};

function problem(status: number, detail: string): Problem {
  return { status, title: titles[status] ?? 'Erro', detail };
}

function problemOf(error: FastifyError): Problem {
  if (error.validation) {
    return problem(400, validationMessage(error.validation, error.validationContext));
  }
  if (error instanceof ValidationError) {
    return problem(400, error.message);
  }
  if (error instanceof NotFoundError) {
    return problem(404, error.message);
  }
  if (error instanceof ConflictError) {
    return problem(409, error.message);
  }
  if (error instanceof RuleViolationError) {
    return problem(422, error.message);
  }
  if (error.statusCode && error.statusCode >= 400 && error.statusCode < 500) {
    return problem(error.statusCode, requestErrors[error.code] ?? error.message);
  }
  return { status: 500, title: titles[500] ?? 'Erro' };
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
    .send(problem(404, `A rota ${request.method} ${request.url.split('?')[0]} não existe.`));
}
