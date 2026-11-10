import type { Repositories } from '@sample-app/db';
import { type Clock, systemClock } from './services/clock';
import { OrderService } from './services/order-service';
// romanctl:imports

export interface Context {
  orders: OrderService;
  // romanctl:services
}

export function createContext(repositories: Repositories, clock: Clock = systemClock): Context {
  return {
    orders: new OrderService(repositories.orders, clock),
    // romanctl:context
  };
}
