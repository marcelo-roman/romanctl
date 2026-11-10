import { QueryClient } from '@tanstack/react-query';
import { isApiError } from '@/shared/api';

const maxRetries = 1;

function shouldRetry(failureCount: number, error: unknown): boolean {
  if (isApiError(error) && error.status >= 400 && error.status < 500) {
    return false;
  }
  return failureCount < maxRetries;
}

export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: { staleTime: 10_000, retry: shouldRetry, refetchOnWindowFocus: false },
    },
  });
}
