import { useQuery } from '@tanstack/react-query';
import { fetchOrder, fetchOrders } from './order-api';

export const recentOrdersLimit = 50;

export const orderKeys = {
  all: ['orders'] as const,
  list: (limit: number) => [...orderKeys.all, 'list', limit] as const,
  detail: (id: string) => [...orderKeys.all, 'detail', id] as const,
};

export function useOrder(id: string) {
  return useQuery({ queryKey: orderKeys.detail(id), queryFn: () => fetchOrder(id) });
}

export function useOrders(limit: number = recentOrdersLimit) {
  return useQuery({ queryKey: orderKeys.list(limit), queryFn: () => fetchOrders(limit) });
}
