import { useMutation, useQueryClient } from '@tanstack/react-query';
import { orderKeys } from '@/entities/order';
import { errorMessage } from '@/shared/api';
import { placeOrder } from '../api/place-order';

export function usePlaceOrder(orderId: string) {
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: () => placeOrder(orderId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: orderKeys.all }),
  });

  return {
    place: () => mutation.mutate(),
    isPending: mutation.isPending,
    error: mutation.isError ? errorMessage(mutation.error) : undefined,
  };
}
