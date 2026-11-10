import { QueryClient } from '@tanstack/react-query';
import { TRPCClientError } from '@trpc/client';

const MAX_RETRIES = 2;

function isClientError(error: unknown): boolean {
  if (!(error instanceof TRPCClientError)) {
    return false;
  }
  const status: unknown = error.data?.httpStatus;
  return typeof status === 'number' && status >= 400 && status < 500;
}

export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: (failures, error) => failures < MAX_RETRIES && !isClientError(error) },
    },
  });
}
