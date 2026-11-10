'use client';

import { useMutation } from '@tanstack/react-query';
import { errorMessage } from '../../lib/error-message';
import { useTRPC } from '../../trpc/react';
import { useOrderUpdated } from './use-order-updated';

interface PlaceOrderButtonProps {
  orderId: string;
}

export function PlaceOrderButton({ orderId }: PlaceOrderButtonProps) {
  const trpc = useTRPC();
  const orderUpdated = useOrderUpdated();
  const place = useMutation(trpc.orders.place.mutationOptions({ onSuccess: orderUpdated }));

  return (
    <div>
      <button type="button" disabled={place.isPending} onClick={() => place.mutate({ orderId })}>
        Realizar pedido
      </button>
      {place.isError && <p role="alert">{errorMessage(place.error)}</p>}
    </div>
  );
}
