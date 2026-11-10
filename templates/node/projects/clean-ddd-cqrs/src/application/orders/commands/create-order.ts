import { Inject } from '@nestjs/common';
import { Command, CommandHandler, type ICommandHandler } from '@nestjs/cqrs';
import { Order } from '../../../domain/orders/order.js';
import { OrderRepository } from '../../../domain/orders/order-repository.js';
import { Clock } from '../../common/clock.js';

export class CreateOrder extends Command<string> {
  constructor(readonly customerName: string) {
    super();
  }
}

@CommandHandler(CreateOrder)
export class CreateOrderHandler implements ICommandHandler<CreateOrder, string> {
  constructor(
    @Inject(OrderRepository) private readonly orders: OrderRepository,
    @Inject(Clock) private readonly clock: Clock,
  ) {}

  async execute(command: CreateOrder): Promise<string> {
    const order = Order.create(command.customerName, this.clock.now());
    await this.orders.save(order);
    return order.id.value;
  }
}
