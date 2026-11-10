import { useQueryClient } from '@tanstack/react-query';
import { type ApiOutputs, useTRPC } from '../../trpc/react';

type Order = ApiOutputs['orders']['byId'];

export function useOrderUpdated(): (order: Order) => Promise<void> {
  const trpc = useTRPC();
  const queryClient = useQueryClient();
  return async (order) => {
    queryClient.setQueryData(trpc.orders.byId.queryKey({ orderId: order.id }), order);
    await queryClient.invalidateQueries({ queryKey: trpc.orders.list.queryKey() });
  };
}
