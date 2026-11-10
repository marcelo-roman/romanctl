import Fastify, { type FastifyInstance } from 'fastify';
import { OrderActions } from './actions/order-actions.js';
import { ApiHttp } from './api/api-http.js';
import { CatalogApi } from './api/catalog-api.js';
import { OrdersApi } from './api/orders-api.js';
import { healthRoutes } from './routes/health.js';
import { orderRoutes } from './routes/orders.js';
import { sendError, sendNotFound } from './routes/problem.js';
import { serveWeb } from './routes/web.js';
import { OrdersScreen } from './screens/orders-screen.js';
// romanctl:imports

export interface AppOptions {
  apiUrl: string;
  webDist?: string;
  logger?: boolean;
}

export function buildApp(options: AppOptions): FastifyInstance {
  const app = Fastify({ logger: options.logger ?? false });
  const http = new ApiHttp(options.apiUrl);
  const orders = new OrdersApi(http);
  const catalog = new CatalogApi(http);
  const ordersScreen = new OrdersScreen(orders, catalog);
  app.setErrorHandler(sendError);
  app.register(healthRoutes);
  app.register(orderRoutes(ordersScreen, new OrderActions(orders, catalog, ordersScreen)));
  // romanctl:routes
  if (options.webDist) {
    serveWeb(app, options.webDist);
  } else {
    app.setNotFoundHandler(sendNotFound);
  }
  return app;
}
