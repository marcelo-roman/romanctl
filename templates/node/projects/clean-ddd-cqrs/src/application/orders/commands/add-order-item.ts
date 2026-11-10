import { Inject } from '@nestjs/common';
import { Command, CommandHandler, type ICommandHandler } from '@nestjs/cqrs';
import { DEFAULT_CURRENCY, Money } from '../../../domain/orders/money.js';
import { OrderId } from '../../../domain/orders/order-id.js';
import { OrderRepository } from '../../../domain/orders/order-repository.js';
import { NotFoundError } from '../../common/not-found-error.js';

export class AddOrderItem extends Command<void> {
  constructor(
    readonly orderId: string,
    readonly productName: string,
    readonly quantity: number,
    readonly unitPrice: number,
    readonly currency: string = DEFAULT_CURRENCY,
  ) {
    super();
  }
}

@CommandHandler(AddOrderItem)
export class AddOrderItemHandler implements ICommandHandler<AddOrderItem, void> {
  constructor(@Inject(OrderRepository) private readonly orders: OrderRepository) {}

  async execute(command: AddOrderItem): Promise<void> {
    const order = await this.orders.findById(OrderId.parse(command.orderId));
    if (!order) {
      throw new NotFoundError(`Pedido ${command.orderId} não encontrado.`);
    }
    order.addItem(command.productName, command.quantity, Money.fromAmount(command.unitPrice, command.currency));
    await this.orders.save(order);
  }
}
