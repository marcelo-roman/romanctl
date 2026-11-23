import type { FastifyError, FastifySchemaValidationError } from 'fastify';

const typeNames: Record<string, string> = {
  string: 'a string',
  boolean: 'true or false',
  integer: 'an integer',
  number: 'a number',
  object: 'an object',
  'string,null': 'a string or null',
};

const formatNames: Record<string, string> = {
  date: 'YYYY-MM-DD',
  'date-time': 'ISO 8601 date and time',
  uuid: 'UUID',
};

function fieldOf(error: FastifySchemaValidationError, context: FastifyError['validationContext']): string {
  const path = error.instancePath.replace(/^\//, '').replaceAll('/', '.');
  return path || context || 'body';
}

function describe(error: FastifySchemaValidationError, field: string): string {
  const { params } = error;
  switch (error.keyword) {
    case 'required':
      return `The field ${String(params.missingProperty)} is required.`;
    case 'type':
      return `The field ${field} must be ${typeNames[String(params.type)] ?? String(params.type)}.`;
    case 'minLength':
      return `The field ${field} cannot be empty.`;
    case 'maxLength':
      return `The field ${field} must be at most ${String(params.limit)} characters.`;
    case 'format':
      return `The field ${field} must be in the format ${formatNames[String(params.format)] ?? String(params.format)}.`;
    case 'minProperties':
      return 'Provide at least one field.';
    case 'additionalProperties':
      return `The field ${String(params.additionalProperty)} is not allowed.`;
  }
  return `The field ${field} is invalid.`;
}

export function validationMessage(errors: FastifySchemaValidationError[], context: FastifyError['validationContext']): string {
  const [first] = errors;
  if (!first) {
    return 'Invalid request.';
  }
  return describe(first, fieldOf(first, context));
}
