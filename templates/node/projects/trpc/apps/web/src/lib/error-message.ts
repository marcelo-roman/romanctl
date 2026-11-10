import { TRPCClientError } from '@trpc/client';

const FALLBACK = 'Algo deu errado. Tente de novo em instantes.';
const INVALID_INPUT = 'Confira os dados informados.';

function issuesOf(data: unknown): string[] {
  if (typeof data !== 'object' || data === null || !('issues' in data) || !Array.isArray(data.issues)) {
    return [];
  }
  return data.issues.filter((issue: unknown): issue is string => typeof issue === 'string');
}

export function errorMessage(error: unknown): string {
  if (!(error instanceof TRPCClientError)) {
    return FALLBACK;
  }
  const data: unknown = error.data;
  const code = typeof data === 'object' && data !== null && 'code' in data ? data.code : undefined;
  switch (code) {
    case 'BAD_REQUEST':
      return issuesOf(data)[0] ?? INVALID_INPUT;
    case 'NOT_FOUND':
    case 'UNPROCESSABLE_CONTENT':
    case 'PRECONDITION_FAILED':
      return error.message;
    default:
      return FALLBACK;
  }
}
