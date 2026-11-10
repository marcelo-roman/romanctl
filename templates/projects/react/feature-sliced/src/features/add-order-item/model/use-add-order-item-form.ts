import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { formatMoney, orderKeys } from '@/entities/order';
import { errorMessage } from '@/shared/api';
import { addOrderItem } from '../api/add-order-item';
import { emptyItemForm, previewSubtotal, toAddItemInput, type ItemFormValues } from './item-form';

export function useAddOrderItemForm(orderId: string) {
  const [values, setValues] = useState<ItemFormValues>(emptyItemForm);
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: () => addOrderItem(orderId, toAddItemInput(values)),
    onSuccess: async () => {
      setValues(emptyItemForm);
      await queryClient.invalidateQueries({ queryKey: orderKeys.all });
    },
  });

  const change = (field: keyof ItemFormValues) => (value: string) => {
    setValues((current) => ({ ...current, [field]: value }));
  };

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    mutation.mutate();
  };

  const subtotal = previewSubtotal(values);

  return {
    values,
    change,
    submit,
    subtotal: subtotal ? formatMoney(subtotal) : undefined,
    isPending: mutation.isPending,
    error: mutation.isError ? errorMessage(mutation.error) : undefined,
  };
}
