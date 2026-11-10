import { Inject } from '@nestjs/common';
import { Command, CommandHandler, EventBus, type ICommandHandler } from '@nestjs/cqrs';
import { OrderId } from '../../../domain/orders/order-id.js';
import { OrderRepository } from '../../../domain/orders/order-repository.js';
import { Clock } from '../../common/clock.js';
import { NotFoundError } from '../../common/not-found-error.js';

export class PlaceOrder extends Command<void> {
  constructor(readonly orderId: string) {
    super();
  }
}

@CommandHandler(PlaceOrder)
export class PlaceOrderHandler implements ICommandHandler<PlaceOrder, void> {
  constructor(
    @Inject(OrderRepository) private readonly orders: OrderRepository,
    @Inject(Clock) private readonly clock: Clock,
    @Inject(EventBus) private readonly events: EventBus,
  ) {}

  async execute(command: PlaceOrder): Promise<void> {
    const order = await this.orders.findById(OrderId.parse(command.orderId));
    if (!order) {
      throw new NotFoundError(`Pedido ${command.orderId} não encontrado.`);
    }
    order.place(this.clock.now());
    await this.orders.save(order);
    this.events.publishAll(order.pullEvents());
  }
}
