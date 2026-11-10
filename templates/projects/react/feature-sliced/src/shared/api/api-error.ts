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
  [unreachableStatus]: 'Não foi possível falar com a API. Confira se ela está no ar.',
  400: 'A requisição tem dados inválidos.',
  404: 'Não encontrado.',
  422: 'A operação não é permitida.',
  500: 'A API encontrou um erro inesperado.',
};

export function defaultTitle(status: number): string {
  return titlesByStatus[status] ?? `A API respondeu com o status ${status}.`;
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}
