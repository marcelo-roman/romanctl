import { Body, Controller, Get, HttpCode, HttpStatus, Inject, Param, Post, Query, Res } from '@nestjs/common';
import { CommandBus, QueryBus } from '@nestjs/cqrs';
import { AddOrderItem } from '../application/orders/commands/add-order-item.js';
import { CreateOrder } from '../application/orders/commands/create-order.js';
import { PlaceOrder } from '../application/orders/commands/place-order.js';
import type { OrderView } from '../application/orders/order-view.js';
import { GetOrderById } from '../application/orders/queries/get-order-by-id.js';
import { ListOrders } from '../application/orders/queries/list-orders.js';
import type { Location } from './location.js';
import { RequestBody } from './request-body.js';

@Controller('orders')
export class OrdersController {
  constructor(
    @Inject(CommandBus) private readonly commands: CommandBus,
    @Inject(QueryBus) private readonly queries: QueryBus,
  ) {}

  @Post()
  async create(@Body() body: unknown, @Res({ passthrough: true }) response: Location): Promise<{ id: string }> {
    const input = RequestBody.from(body);
    const id = await this.commands.execute(new CreateOrder(input.text('customerName')));
    response.location(`/orders/${id}`);
    return { id };
  }

  @Post(':id/items')
  @HttpCode(HttpStatus.NO_CONTENT)
  async addItem(@Param('id') id: string, @Body() body: unknown): Promise<void> {
    const input = RequestBody.from(body);
    await this.commands.execute(
      new AddOrderItem(id, input.text('productName'), input.integer('quantity'), input.number('unitPrice'), input.optionalText('currency')),
    );
  }

  @Post(':id/place')
  @HttpCode(HttpStatus.NO_CONTENT)
  async place(@Param('id') id: string): Promise<void> {
    await this.commands.execute(new PlaceOrder(id));
  }

  @Get(':id')
  get(@Param('id') id: string): Promise<OrderView> {
    return this.queries.execute(new GetOrderById(id));
  }

  @Get()
  list(@Query('limit') limit?: string): Promise<OrderView[]> {
    return this.queries.execute(new ListOrders(limit ? Number(limit) : undefined));
  }
}
