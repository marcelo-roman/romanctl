import type { AppRouter } from '@sample-app/api';
import type { inferRouterOutputs } from '@trpc/server';
import { createTRPCContext } from '@trpc/tanstack-react-query';

export const { TRPCProvider, useTRPC } = createTRPCContext<AppRouter>();

export type ApiOutputs = inferRouterOutputs<AppRouter>;
