import { getJson } from '@/shared/api';
import type { Order } from '../model/order';
import type { OrderDto } from './order-dto';
import { orderOf } from './order-mapper';

export async function fetchOrder(id: string): Promise<Order> {
  return orderOf(await getJson<OrderDto>(`/orders/${encodeURIComponent(id)}`));
}

export async function fetchOrders(limit: number): Promise<Order[]> {
  const dtos = await getJson<OrderDto[]>(`/orders?limit=${limit}`);
  return dtos.map(orderOf);
}
