export class ValidationError extends Error {
  override readonly name = 'ValidationError';
}

export class NotFoundError extends Error {
  override readonly name = 'NotFoundError';
}

export class ConflictError extends Error {
  override readonly name = 'ConflictError';
}

export class RuleViolationError extends Error {
  override readonly name = 'RuleViolationError';
}
