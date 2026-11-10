import { TRPCClientError } from '@trpc/client';
import { describe, expect, it } from 'vitest';
import { errorMessage } from '../src/lib/error-message';
import { parseDecimal } from '../src/lib/format';

function clientError(message: string, data: unknown) {
  return new TRPCClientError(message, { result: { error: { message, code: -32600, data } } });
}

describe('errorMessage', () => {
  it('shows the first validation issue for BAD_REQUEST', () => {
    expect(errorMessage(clientError('x', { code: 'BAD_REQUEST', issues: ['Informe o nome do cliente.'] }))).toBe('Informe o nome do cliente.');
    expect(errorMessage(clientError('x', { code: 'BAD_REQUEST', issues: [] }))).toBe('Confira os dados informados.');
  });

  it('shows the server message for rules and missing orders', () => {
    for (const code of ['UNPROCESSABLE_CONTENT', 'PRECONDITION_FAILED', 'NOT_FOUND']) {
      expect(errorMessage(clientError('Mensagem da regra.', { code }))).toBe('Mensagem da regra.');
    }
  });

  it('hides anything else behind a generic message', () => {
    expect(errorMessage(clientError('stack trace', { code: 'INTERNAL_SERVER_ERROR' }))).toMatch(/Algo deu errado/);
    expect(errorMessage(new Error('boom'))).toMatch(/Algo deu errado/);
  });
});

describe('parseDecimal', () => {
  it('accepts a comma as the decimal separator and treats blank as not a number', () => {
    expect(parseDecimal(' 3,50 ')).toBe(3.5);
    expect(parseDecimal('')).toBeNaN();
  });
});
