import type { FastifyError, FastifySchemaValidationError } from 'fastify';

const typeNames: Record<string, string> = {
  string: 'um texto',
  boolean: 'verdadeiro ou falso',
  integer: 'um número inteiro',
  number: 'um número',
  object: 'um objeto',
  'string,null': 'um texto ou null',
};

const formatNames: Record<string, string> = {
  date: 'AAAA-MM-DD',
  'date-time': 'data e hora ISO 8601',
  uuid: 'UUID',
};

function fieldOf(error: FastifySchemaValidationError, context: FastifyError['validationContext']): string {
  const path = error.instancePath.replace(/^\//, '').replaceAll('/', '.');
  return path || context || 'corpo';
}

function describe(error: FastifySchemaValidationError, field: string): string {
  const { params } = error;
  switch (error.keyword) {
    case 'required':
      return `O campo ${String(params.missingProperty)} é obrigatório.`;
    case 'type':
      return `O campo ${field} deve ser ${typeNames[String(params.type)] ?? String(params.type)}.`;
    case 'minLength':
      return `O campo ${field} não pode ficar vazio.`;
    case 'maxLength':
      return `O campo ${field} deve ter no máximo ${String(params.limit)} caracteres.`;
    case 'format':
      return `O campo ${field} deve estar no formato ${formatNames[String(params.format)] ?? String(params.format)}.`;
    case 'minProperties':
      return 'Informe ao menos um campo.';
    case 'additionalProperties':
      return `O campo ${String(params.additionalProperty)} não é aceito.`;
  }
  return `O campo ${field} é inválido.`;
}

export function validationMessage(errors: FastifySchemaValidationError[], context: FastifyError['validationContext']): string {
  const [first] = errors;
  if (!first) {
    return 'Requisição inválida.';
  }
  return describe(first, fieldOf(first, context));
}
