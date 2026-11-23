import type { FastifySchemaValidationError } from 'fastify';

const fieldNames: Record<string, string> = {
  customerName: 'customer name',
  productId: 'product',
  quantity: 'quantity',
  name: 'name',
};

const typeNames: Record<string, string> = {
  string: 'a string',
  integer: 'a whole number',
  number: 'a number',
};

function fieldName(field: string): string {
  return fieldNames[field] ?? field;
}

export function validationMessage(errors: FastifySchemaValidationError[]): string {
  const [first] = errors;
  if (!first) {
    return 'bad request.';
  }
  const field = fieldName(first.instancePath.replace(/^\//, ''));
  switch (first.keyword) {
    case 'required':
      return `fill in the ${fieldName(String(first.params.missingProperty))} field.`;
    case 'type':
      return `the ${field} field must be ${typeNames[String(first.params.type)] ?? String(first.params.type)}.`;
    case 'minLength':
      return `the ${field} field cannot be empty.`;
  }
  return `the ${field} field is invalid.`;
}
