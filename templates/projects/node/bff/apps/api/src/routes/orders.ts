import type { FastifyPluginAsync } from 'fastify';
import type { AddOrderItemInput, CreateOrderInput, ListOrdersInput, OrderService } from '../services/order-service.js';
import { addOrderItemSchema, createOrderSchema, getOrderSchema, listOrdersSchema, placeOrderSchema } from './order-schemas.js';
import { orderView } from './views.js';

interface IdParams {
  id: string;
}

export function orderRoutes(service: OrderService): FastifyPluginAsync {
  return async (app) => {
    app.post<{ Body: CreateOrderInput }>('/orders', { schema: createOrderSchema }, async (request, reply) => {
      const order = await service.create(request.body);
      return reply.status(201).header('location', `/orders/${order.id}`).send(orderView(order));
    });

    app.get<{ Querystring: ListOrdersInput }>('/orders', { schema: listOrdersSchema }, async (request) =>
      (await service.list(request.query)).map(orderView),
    );

    app.get<{ Params: IdParams }>('/orders/:id', { schema: getOrderSchema }, async (request) => orderView(await service.get(request.params.id)));

    app.post<{ Params: IdParams; Body: AddOrderItemInput }>('/orders/:id/items', { schema: addOrderItemSchema }, async (request) =>
      orderView(await service.addItem(request.params.id, request.body)),
    );

    app.post<{ Params: IdParams }>('/orders/:id/place', { schema: placeOrderSchema }, async (request) =>
      orderView(await service.place(request.params.id)),
    );
  };
}
