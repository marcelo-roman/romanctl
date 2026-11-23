'use client';

import { useQuery } from '@tanstack/react-query';
import { errorMessage } from '../../lib/error-message';
import { formatMoney, statusLabel } from '../../lib/format';
import { useTRPC } from '../../trpc/react';
import { AddItemForm } from './add-item-form';
import { PlaceOrderButton } from './place-order-button';

interface OrderDetailsProps {
  orderId: string;
}

export function OrderDetails({ orderId }: OrderDetailsProps) {
  const trpc = useTRPC();
  const order = useQuery(trpc.orders.byId.queryOptions({ orderId }));

  if (order.isPending) {
    return <p>Loading order...</p>;
  }
  if (order.isError) {
    return <p role="alert">{errorMessage(order.error)}</p>;
  }
  const { customerName, status, items, total, currency } = order.data;
  return (
    <section>
      <h1>Order for {customerName}</h1>
      <p>
        Status: <strong>{statusLabel(status)}</strong>
      </p>
      {items.length === 0 ? (
        <p>No items yet.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Product</th>
              <th>Quantity</th>
              <th>Unit price</th>
              <th>Subtotal</th>
            </tr>
          </thead>
          <tbody>
            {items.map((item) => (
              <tr key={item.id}>
                <td>{item.productName}</td>
                <td>{item.quantity}</td>
                <td>{formatMoney(item.unitPrice, currency)}</td>
                <td>{formatMoney(item.subtotal, currency)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      <p>
        Total: <strong data-testid="order-total">{formatMoney(total, currency)}</strong>
      </p>
      <AddItemForm orderId={orderId} />
      <PlaceOrderButton orderId={orderId} />
    </section>
  );
}
