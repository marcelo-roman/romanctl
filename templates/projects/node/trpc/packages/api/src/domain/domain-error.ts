export class DomainError extends Error {
  override readonly name: string = 'DomainError';
}

export class InvalidStateError extends DomainError {
  override readonly name: string = 'InvalidStateError';
}
