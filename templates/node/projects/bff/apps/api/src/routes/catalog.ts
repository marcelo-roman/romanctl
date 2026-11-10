import type { FastifyPluginAsync } from 'fastify';
import type { AddCatalogProductInput, CatalogService } from '../services/catalog-service.js';
import { addCatalogProductSchema, getCatalogProductSchema, listCatalogProductsSchema } from './catalog-schemas.js';
import { catalogProductView } from './views.js';

interface IdParams {
  id: string;
}

export function catalogRoutes(service: CatalogService): FastifyPluginAsync {
  return async (app) => {
    app.post<{ Body: AddCatalogProductInput }>('/catalog/products', { schema: addCatalogProductSchema }, async (request, reply) => {
      const product = await service.add(request.body);
      return reply.status(201).header('location', `/catalog/products/${product.id}`).send(catalogProductView(product));
    });

    app.get('/catalog/products', { schema: listCatalogProductsSchema }, async () => (await service.list()).map(catalogProductView));

    app.get<{ Params: IdParams }>('/catalog/products/:id', { schema: getCatalogProductSchema }, async (request) =>
      catalogProductView(await service.get(request.params.id)),
    );
  };
}
