import type { FastifyPluginAsync } from 'fastify';
import type { AddItemAction, CreateOrderAction, OrderActions } from '../actions/order-actions.js';
import type { OrdersScreen } from '../screens/orders-screen.js';

interface IdParams {
  id: string;
}

const createOrderSchema = {
  body: {
    type: 'object',
    required: ['customerName'],
    properties: { customerName: { type: 'string' } },
  },
} as const;

const addItemSchema = {
  body: {
    type: 'object',
    required: ['productId', 'quantity'],
    properties: { productId: { type: 'string', minLength: 1 }, quantity: { type: 'integer' } },
  },
} as const;

export function orderRoutes(screen: OrdersScreen, actions: OrderActions): FastifyPluginAsync {
  return async (app) => {
    app.get('/screens/orders', () => screen.list());

    app.get<{ Params: IdParams }>('/screens/orders/:id', (request) => screen.detail(request.params.id));

    app.post<{ Body: CreateOrderAction }>('/actions/orders', { schema: createOrderSchema }, async (request, reply) =>
      reply.status(201).send(await actions.create(request.body)),
    );

    app.post<{ Params: IdParams; Body: AddItemAction }>('/actions/orders/:id/items', { schema: addItemSchema }, (request) =>
      actions.addItem(request.params.id, request.body),
    );

    app.post<{ Params: IdParams }>('/actions/orders/:id/place', (request) => actions.place(request.params.id));
  };
}
