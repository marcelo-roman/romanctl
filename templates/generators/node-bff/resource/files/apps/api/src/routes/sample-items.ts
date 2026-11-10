import type { FastifyPluginAsync } from 'fastify';
import type { SampleItem } from '../domain/sample-item.js';
import type { ListSampleItemsInput, SampleItemInput, SampleItemService } from '../services/sample-item-service.js';
import {
  createSampleItemSchema,
  deleteSampleItemSchema,
  getSampleItemSchema,
  listSampleItemsSchema,
  renameSampleItemSchema,
} from './sample-item-schemas.js';

interface IdParams {
  id: string;
}

function sampleItemView(sampleItem: SampleItem) {
  return { id: sampleItem.id, name: sampleItem.name, createdAt: sampleItem.createdAt.toISOString() };
}

export function sampleItemRoutes(service: SampleItemService): FastifyPluginAsync {
  return async (app) => {
    app.post<{ Body: SampleItemInput }>('/sample-items', { schema: createSampleItemSchema }, async (request, reply) => {
      const sampleItem = await service.create(request.body);
      return reply.status(201).header('location', `/sample-items/${sampleItem.id}`).send(sampleItemView(sampleItem));
    });

    app.get<{ Querystring: ListSampleItemsInput }>('/sample-items', { schema: listSampleItemsSchema }, async (request) =>
      (await service.list(request.query)).map(sampleItemView),
    );

    app.get<{ Params: IdParams }>('/sample-items/:id', { schema: getSampleItemSchema }, async (request) =>
      sampleItemView(await service.get(request.params.id)),
    );

    app.put<{ Params: IdParams; Body: SampleItemInput }>('/sample-items/:id', { schema: renameSampleItemSchema }, async (request) =>
      sampleItemView(await service.rename(request.params.id, request.body)),
    );

    app.delete<{ Params: IdParams }>('/sample-items/:id', { schema: deleteSampleItemSchema }, async (request, reply) => {
      await service.delete(request.params.id);
      return reply.status(204).send();
    });
  };
}
