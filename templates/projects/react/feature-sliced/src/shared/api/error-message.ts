import { isApiError } from './api-error';

const unexpected = 'Algo deu errado. Tente de novo.';

export function errorMessage(error: unknown): string {
  if (!isApiError(error)) {
    return unexpected;
  }
  return error.detail || error.title;
}
