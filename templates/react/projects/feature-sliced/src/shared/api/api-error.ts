export interface Problem {
  status: number;
  title: string;
  detail?: string;
}

export class ApiError extends Error {
  readonly status: number;
  readonly title: string;
  readonly detail?: string;

  constructor(problem: Problem) {
    super(problem.detail || problem.title);
    this.name = 'ApiError';
    this.status = problem.status;
    this.title = problem.title;
    this.detail = problem.detail;
  }
}

export const unreachableStatus = 0;

const titlesByStatus: Record<number, string> = {
  [unreachableStatus]: 'Could not reach the API. Check that it is running.',
  400: 'The request has invalid data.',
  404: 'Not found.',
  422: 'The operation is not allowed.',
  500: 'The API hit an unexpected error.',
};

export function defaultTitle(status: number): string {
  return titlesByStatus[status] ?? `The API answered with status ${status}.`;
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}
