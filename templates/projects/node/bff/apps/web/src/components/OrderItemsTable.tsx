import type { OrderLine } from '../bff/orders';

interface OrderItemsTableProps {
  items: OrderLine[];
  total: string;
}

export function OrderItemsTable({ items, total }: OrderItemsTableProps) {
  if (items.length === 0) {
    return <p>O pedido ainda não tem itens.</p>;
  }
  return (
    <table>
      <thead>
        <tr>
          <th>Produto</th>
          <th>Quantidade</th>
          <th>Preço</th>
          <th>Subtotal</th>
        </tr>
      </thead>
      <tbody>
        {items.map((item) => (
          <tr key={item.id}>
            <td>{item.productName}</td>
            <td>{item.quantity}</td>
            <td>{item.unitPrice}</td>
            <td>{item.subtotal}</td>
          </tr>
        ))}
      </tbody>
      <tfoot>
        <tr>
          <td colSpan={3}>Total</td>
          <td>{total}</td>
        </tr>
      </tfoot>
    </table>
  );
}
