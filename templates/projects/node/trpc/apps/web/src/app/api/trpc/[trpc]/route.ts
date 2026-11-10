import { appRouter } from '@sample-app/api';
import { fetchRequestHandler } from '@trpc/server/adapters/fetch';
import { createRequestContext } from '../../../../server/context';

function handler(request: Request): Promise<Response> {
  return fetchRequestHandler({ endpoint: '/api/trpc', req: request, router: appRouter, createContext: createRequestContext });
}

export { handler as GET, handler as POST };
