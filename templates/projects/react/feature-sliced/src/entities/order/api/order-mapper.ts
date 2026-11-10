import { money } from '../model/money';
import type { Order, OrderItem, OrderStatus } from '../model/order';
import type { OrderDto, OrderItemDto } from './order-dto';

function statusOf(status: string): OrderStatus {
  return status === 'Placed' ? 'Placed' : 'Draft';
}

function itemOf(dto: OrderItemDto, currency: string): OrderItem {
  return {
    productName: dto.productName,
    quantity: dto.quantity,
    unitPrice: money(dto.unitPrice, currency),
    subtotal: money(dto.subtotal, currency),
  };
}

export function orderOf(dto: OrderDto): Order {
  return {
    id: dto.id,
    customerName: dto.customerName,
    status: statusOf(dto.status),
    createdAt: new Date(dto.createdAt),
    total: money(dto.total, dto.currency),
    items: dto.items.map((item) => itemOf(item, dto.currency)),
  };
}
