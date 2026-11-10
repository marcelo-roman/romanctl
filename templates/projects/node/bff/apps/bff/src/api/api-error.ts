export class ApiError extends Error {
  override readonly name = 'ApiError';

  constructor(
    readonly status: number,
    readonly detail: string,
  ) {
    super(detail);
  }
}

export class ApiUnavailableError extends Error {
  override readonly name = 'ApiUnavailableError';
}
