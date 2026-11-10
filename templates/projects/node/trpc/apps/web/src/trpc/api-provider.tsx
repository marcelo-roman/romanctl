'use client';

import { QueryClientProvider } from '@tanstack/react-query';
import { type ReactNode, useState } from 'react';
import { type ApiClient, createApiClient } from './client';
import { createQueryClient } from './query-client';
import { TRPCProvider } from './react';

interface ApiProviderProps {
  children: ReactNode;
  client?: ApiClient;
}

export function ApiProvider({ children, client }: ApiProviderProps) {
  const [queryClient] = useState(createQueryClient);
  const [trpcClient] = useState(() => client ?? createApiClient());
  return (
    <QueryClientProvider client={queryClient}>
      <TRPCProvider trpcClient={trpcClient} queryClient={queryClient}>
        {children}
      </TRPCProvider>
    </QueryClientProvider>
  );
}
