import { Inject } from '@nestjs/common';
import { type IQueryHandler, Query, QueryHandler } from '@nestjs/cqrs';
import { listLimit } from '../../common/list-limit.js';
import { OrderReadStore } from '../order-read-store.js';
import type { OrderView } from '../order-view.js';

export class ListOrders extends Query<OrderView[]> {
  constructor(readonly limit?: number) {
    super();
  }
}

@QueryHandler(ListOrders)
export class ListOrdersHandler implements IQueryHandler<ListOrders, OrderView[]> {
  constructor(@Inject(OrderReadStore) private readonly store: OrderReadStore) {}

  execute(query: ListOrders): Promise<OrderView[]> {
    return this.store.list(listLimit(query.limit));
  }
}
