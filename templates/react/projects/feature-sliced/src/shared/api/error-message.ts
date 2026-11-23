import { isApiError } from './api-error';

const unexpected = 'Something went wrong. Try again.';

export function errorMessage(error: unknown): string {
  if (!isApiError(error)) {
    return unexpected;
  }
  return error.detail || error.title;
}
