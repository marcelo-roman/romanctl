import { TRPCError } from '@trpc/server';
import { DomainError, InvalidStateError } from '../domain/domain-error';
import { NotFoundError } from '../services/not-found-error';

export function toTRPCError(error: TRPCError): TRPCError {
  const cause = error.cause;
  if (cause instanceof InvalidStateError) {
    return new TRPCError({ code: 'PRECONDITION_FAILED', message: cause.message, cause });
  }
  if (cause instanceof DomainError) {
    return new TRPCError({ code: 'UNPROCESSABLE_CONTENT', message: cause.message, cause });
  }
  if (cause instanceof NotFoundError) {
    return new TRPCError({ code: 'NOT_FOUND', message: cause.message, cause });
  }
  return error;
}

export function validationIssues(error: TRPCError): string[] {
  const cause: unknown = error.cause;
  if (error.code !== 'BAD_REQUEST' || typeof cause !== 'object' || cause === null || !('issues' in cause)) {
    return [];
  }
  const issues: unknown = cause.issues;
  if (!Array.isArray(issues)) {
    return [];
  }
  return issues.flatMap((issue: unknown) =>
    typeof issue === 'object' && issue !== null && 'message' in issue && typeof issue.message === 'string' ? [issue.message] : [],
  );
}
