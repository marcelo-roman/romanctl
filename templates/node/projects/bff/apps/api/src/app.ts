import Fastify, { type FastifyInstance } from 'fastify';
import type { Repositories } from './repositories/repositories.js';
import { catalogRoutes } from './routes/catalog.js';
import { healthRoutes } from './routes/health.js';
import { orderRoutes } from './routes/orders.js';
import { sendNotFound, sendProblem } from './routes/problem.js';
import { CatalogService } from './services/catalog-service.js';
import { type Clock, systemClock } from './services/clock.js';
import { OrderService } from './services/order-service.js';
// romanctl:imports

export interface AppOptions {
  logger?: boolean;
  clock?: Clock;
}

export function buildApp(repositories: Repositories, options: AppOptions = {}): FastifyInstance {
  const clock = options.clock ?? systemClock;
  const app = Fastify({ logger: options.logger ?? false });
  app.setErrorHandler(sendProblem);
  app.setNotFoundHandler(sendNotFound);
  app.register(healthRoutes);
  app.register(orderRoutes(new OrderService(repositories.orders, clock)));
  app.register(catalogRoutes(new CatalogService(repositories.catalog, clock)));
  // romanctl:routes
  return app;
}
