import { Inject } from '@nestjs/common';
import { type IQueryHandler, Query, QueryHandler } from '@nestjs/cqrs';
import { OrderId } from '../../../domain/orders/order-id.js';
import { NotFoundError } from '../../common/not-found-error.js';
import { OrderReadStore } from '../order-read-store.js';
import type { OrderView } from '../order-view.js';

export class GetOrderById extends Query<OrderView> {
  constructor(readonly orderId: string) {
    super();
  }
}

@QueryHandler(GetOrderById)
export class GetOrderByIdHandler implements IQueryHandler<GetOrderById, OrderView> {
  constructor(@Inject(OrderReadStore) private readonly store: OrderReadStore) {}

  async execute(query: GetOrderById): Promise<OrderView> {
    const order = await this.store.findById(OrderId.parse(query.orderId).value);
    if (!order) {
      throw new NotFoundError(`Pedido ${query.orderId} não encontrado.`);
    }
    return order;
  }
}
