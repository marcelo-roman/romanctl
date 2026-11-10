import Fastify, { type FastifyInstance } from 'fastify';

export interface ReceivedCall {
  method: string;
  url: string;
  body: unknown;
}

export interface FakeApi {
  url: string;
  calls: ReceivedCall[];
  close(): Promise<void>;
}

export async function startFakeApi(routes: (app: FastifyInstance) => void): Promise<FakeApi> {
  const app = Fastify();
  const calls: ReceivedCall[] = [];
  app.addHook('preHandler', async (request) => {
    calls.push({ method: request.method, url: request.url, body: request.body });
  });
  routes(app);
  const url = await app.listen({ port: 0, host: '127.0.0.1' });
  return { url, calls, close: () => app.close() };
}

export function problem(status: number, detail: string) {
  return { status, title: 'Erro', detail };
}
