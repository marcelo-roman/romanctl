'use client';

import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { errorMessage } from '../../lib/error-message';
import { formatDate, formatMoney, statusLabel } from '../../lib/format';
import { useTRPC } from '../../trpc/react';

export function OrderList() {
  const trpc = useTRPC();
  const orders = useQuery(trpc.orders.list.queryOptions());

  if (orders.isPending) {
    return <p>Loading orders...</p>;
  }
  if (orders.isError) {
    return <p role="alert">{errorMessage(orders.error)}</p>;
  }
  if (orders.data.length === 0) {
    return <p>No orders yet.</p>;
  }
  return (
    <table>
      <thead>
        <tr>
          <th>Customer</th>
          <th>Status</th>
          <th>Total</th>
          <th>Created at</th>
        </tr>
      </thead>
      <tbody>
        {orders.data.map((order) => (
          <tr key={order.id}>
            <td>
              <Link href={`/orders/${order.id}`}>{order.customerName}</Link>
            </td>
            <td>{statusLabel(order.status)}</td>
            <td>{formatMoney(order.total, order.currency)}</td>
            <td>{formatDate(order.createdAt)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
