import type { FastifySchemaValidationError } from 'fastify';

const fieldNames: Record<string, string> = {
  customerName: 'nome do cliente',
  productId: 'produto',
  quantity: 'quantidade',
  name: 'nome',
};

const typeNames: Record<string, string> = {
  string: 'um texto',
  integer: 'um número inteiro',
  number: 'um número',
};

function fieldName(field: string): string {
  return fieldNames[field] ?? field;
}

export function validationMessage(errors: FastifySchemaValidationError[]): string {
  const [first] = errors;
  if (!first) {
    return 'requisição inválida.';
  }
  const field = fieldName(first.instancePath.replace(/^\//, ''));
  switch (first.keyword) {
    case 'required':
      return `informe o campo ${fieldName(String(first.params.missingProperty))}.`;
    case 'type':
      return `o campo ${field} deve ser ${typeNames[String(first.params.type)] ?? String(first.params.type)}.`;
    case 'minLength':
      return `o campo ${field} não pode ficar vazio.`;
  }
  return `o campo ${field} é inválido.`;
}
