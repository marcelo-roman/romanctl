import type { OrderRow } from '../bff/orders';
import { Link } from '../router';

interface OrdersTableProps {
  orders: OrderRow[];
}

export function OrdersTable({ orders }: OrdersTableProps) {
  if (orders.length === 0) {
    return <p>Nenhum pedido ainda.</p>;
  }
  return (
    <table>
      <thead>
        <tr>
          <th>Cliente</th>
          <th>Situação</th>
          <th>Itens</th>
          <th>Total</th>
        </tr>
      </thead>
      <tbody>
        {orders.map((order) => (
          <tr key={order.id}>
            <td>
              <Link to={`/orders/${order.id}`}>{order.customerName}</Link>
            </td>
            <td>{order.status}</td>
            <td>{order.itemCount}</td>
            <td>{order.total}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
