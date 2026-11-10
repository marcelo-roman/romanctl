import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { generatePath, useNavigate } from 'react-router';
import { orderKeys } from '@/entities/order';
import { errorMessage } from '@/shared/api';
import { paths } from '@/shared/config';
import { createOrder } from '../api/create-order';

export function useCreateOrderForm() {
  const [customerName, setCustomerName] = useState('');
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const mutation = useMutation({
    mutationFn: createOrder,
    onSuccess: async (id) => {
      await queryClient.invalidateQueries({ queryKey: orderKeys.all });
      await navigate(generatePath(paths.orderDetails, { id }));
    },
  });

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    mutation.mutate(customerName.trim());
  };

  return {
    customerName,
    setCustomerName,
    submit,
    isPending: mutation.isPending,
    error: mutation.isError ? errorMessage(mutation.error) : undefined,
  };
}
