import { Logger } from '@nestjs/common';
import { EventsHandler, type IEventHandler } from '@nestjs/cqrs';
import { OrderPlaced } from '../../../domain/orders/order-placed.js';

@EventsHandler(OrderPlaced)
export class LogOrderPlaced implements IEventHandler<OrderPlaced> {
  private readonly logger = new Logger(LogOrderPlaced.name);

  handle(event: OrderPlaced): void {
    this.logger.log(`Pedido ${event.orderId} realizado: ${event.total} ${event.currency}`);
  }
}
