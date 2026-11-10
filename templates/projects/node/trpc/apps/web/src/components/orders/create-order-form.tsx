'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import type { FormEvent } from 'react';
import { errorMessage } from '../../lib/error-message';
import { useTRPC } from '../../trpc/react';

export function CreateOrderForm() {
  const trpc = useTRPC();
  const queryClient = useQueryClient();
  const router = useRouter();
  const create = useMutation(
    trpc.orders.create.mutationOptions({
      onSuccess: async (order) => {
        await queryClient.invalidateQueries({ queryKey: trpc.orders.list.queryKey() });
        router.push(`/orders/${order.id}`);
      },
    }),
  );

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    create.mutate({ customerName: String(form.get('customerName') ?? '') });
  }

  return (
    <form onSubmit={submit} aria-label="Novo pedido">
      <label>
        Cliente
        <input name="customerName" autoComplete="off" />
      </label>
      <button type="submit" disabled={create.isPending}>
        Criar pedido
      </button>
      {create.isError && <p role="alert">{errorMessage(create.error)}</p>}
    </form>
  );
}
