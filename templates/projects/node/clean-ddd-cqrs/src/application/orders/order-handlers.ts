import { AddOrderItemHandler } from './commands/add-order-item.js';
import { CreateOrderHandler } from './commands/create-order.js';
import { PlaceOrderHandler } from './commands/place-order.js';
import { LogOrderPlaced } from './events/log-order-placed.js';
import { GetOrderByIdHandler } from './queries/get-order-by-id.js';
import { ListOrdersHandler } from './queries/list-orders.js';

export const orderHandlers = [
  CreateOrderHandler,
  AddOrderItemHandler,
  PlaceOrderHandler,
  GetOrderByIdHandler,
  ListOrdersHandler,
  LogOrderPlaced,
];
