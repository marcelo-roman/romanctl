import { TRPCClientError } from '@trpc/client';
import { describe, expect, it } from 'vitest';
import { errorMessage } from '../src/lib/error-message';
import { parseDecimal } from '../src/lib/format';

function clientError(message: string, data: unknown) {
  return new TRPCClientError(message, { result: { error: { message, code: -32600, data } } });
}

describe('errorMessage', () => {
  it('shows the first validation issue for BAD_REQUEST', () => {
    expect(errorMessage(clientError('x', { code: 'BAD_REQUEST', issues: ['Enter the customer name.'] }))).toBe('Enter the customer name.');
    expect(errorMessage(clientError('x', { code: 'BAD_REQUEST', issues: [] }))).toBe('Check the data you entered.');
  });

  it('shows the server message for rules and missing orders', () => {
    for (const code of ['UNPROCESSABLE_CONTENT', 'PRECONDITION_FAILED', 'NOT_FOUND']) {
      expect(errorMessage(clientError('Rule message.', { code }))).toBe('Rule message.');
    }
  });

  it('hides anything else behind a generic message', () => {
    expect(errorMessage(clientError('stack trace', { code: 'INTERNAL_SERVER_ERROR' }))).toMatch(/Something went wrong/);
    expect(errorMessage(new Error('boom'))).toMatch(/Something went wrong/);
  });
});

describe('parseDecimal', () => {
  it('accepts a dot or a comma as the decimal separator and treats blank as not a number', () => {
    expect(parseDecimal(' 3.50 ')).toBe(3.5);
    expect(parseDecimal('3,50')).toBe(3.5);
    expect(parseDecimal('')).toBeNaN();
  });
});
