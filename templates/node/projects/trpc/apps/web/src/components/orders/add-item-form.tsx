'use client';

import { useMutation } from '@tanstack/react-query';
import type { FormEvent } from 'react';
import { errorMessage } from '../../lib/error-message';
import { parseDecimal } from '../../lib/format';
import { useTRPC } from '../../trpc/react';
import { useOrderUpdated } from './use-order-updated';

interface AddItemFormProps {
  orderId: string;
}

export function AddItemForm({ orderId }: AddItemFormProps) {
  const trpc = useTRPC();
  const orderUpdated = useOrderUpdated();
  const addItem = useMutation(trpc.orders.addItem.mutationOptions({ onSuccess: orderUpdated }));

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formElement = event.currentTarget;
    const form = new FormData(formElement);
    addItem.mutate(
      {
        orderId,
        productName: String(form.get('productName') ?? ''),
        quantity: parseDecimal(String(form.get('quantity') ?? '')),
        unitPrice: parseDecimal(String(form.get('unitPrice') ?? '')),
        currency: String(form.get('currency') ?? '') || undefined,
      },
      { onSuccess: () => formElement.reset() },
    );
  }

  return (
    <form onSubmit={submit} aria-label="Add item">
      <label>
        Product
        <input name="productName" autoComplete="off" />
      </label>
      <label>
        Quantity
        <input name="quantity" inputMode="numeric" defaultValue="1" />
      </label>
      <label>
        Unit price
        <input name="unitPrice" inputMode="decimal" />
      </label>
      <label>
        Currency
        <input name="currency" defaultValue="USD" maxLength={3} />
      </label>
      <button type="submit" disabled={addItem.isPending}>
        Add item
      </button>
      {addItem.isError && <p role="alert">{errorMessage(addItem.error)}</p>}
    </form>
  );
}
