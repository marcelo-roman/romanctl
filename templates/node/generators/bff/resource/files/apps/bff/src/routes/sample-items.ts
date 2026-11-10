import type { FastifyPluginAsync } from 'fastify';
import type { SampleItemsScreens } from '../screens/sample-items-screen.js';

interface IdParams {
  id: string;
}

interface NameBody {
  name: string;
}

const nameSchema = {
  body: {
    type: 'object',
    required: ['name'],
    properties: { name: { type: 'string' } },
  },
} as const;

export function sampleItemRoutes(screens: SampleItemsScreens): FastifyPluginAsync {
  return async (app) => {
    app.get('/screens/sample-items', () => screens.list());

    app.post<{ Body: NameBody }>('/actions/sample-items', { schema: nameSchema }, (request) => screens.create(request.body.name));

    app.put<{ Params: IdParams; Body: NameBody }>('/actions/sample-items/:id', { schema: nameSchema }, (request) =>
      screens.rename(request.params.id, request.body.name),
    );

    app.delete<{ Params: IdParams }>('/actions/sample-items/:id', (request) => screens.delete(request.params.id));
  };
}
