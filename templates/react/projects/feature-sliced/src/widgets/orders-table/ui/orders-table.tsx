import { Link } from 'react-router';
import { OrderStatusBadge, useOrders } from '@/entities/order';
import { errorMessage } from '@/shared/api';
import { Alert } from '@/shared/ui';
import { orderRowOf, type OrderRow } from '../model/order-row';

function OrderLine({ row }: { row: OrderRow }) {
  return (
    <tr>
      <td>
        <Link to={row.href}>{row.customerName}</Link>
      </td>
      <td>
        <OrderStatusBadge status={row.status} />
      </td>
      <td>{row.createdAt}</td>
      <td className="numeric">{row.items}</td>
      <td className="numeric">{row.total}</td>
    </tr>
  );
}

export function OrdersTable() {
  const { data, error, isPending } = useOrders();

  if (isPending) {
    return <p className="muted">Carregando pedidos...</p>;
  }
  if (error) {
    return <Alert tone="error">{errorMessage(error)}</Alert>;
  }
  if (data.length === 0) {
    return <p className="muted">Nenhum pedido ainda. Crie o primeiro acima.</p>;
  }
  return (
    <table className="table">
      <thead>
        <tr>
          <th>Cliente</th>
          <th>Status</th>
          <th>Criado em</th>
          <th className="numeric">Itens</th>
          <th className="numeric">Total</th>
        </tr>
      </thead>
      <tbody>
        {data.map(orderRowOf).map((row) => (
          <OrderLine key={row.id} row={row} />
        ))}
      </tbody>
    </table>
  );
}
