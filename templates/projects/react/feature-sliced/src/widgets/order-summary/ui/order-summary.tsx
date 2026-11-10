import { OrderStatusBadge, type Order } from '@/entities/order';
import { summaryOf, type SummaryLine } from '../model/summary';

function ItemLine({ line }: { line: SummaryLine }) {
  return (
    <tr>
      <td>{line.productName}</td>
      <td className="numeric">{line.quantity}</td>
      <td className="numeric">{line.unitPrice}</td>
      <td className="numeric">{line.subtotal}</td>
    </tr>
  );
}

export function OrderSummary({ order }: { order: Order }) {
  const summary = summaryOf(order);

  return (
    <section className="card" aria-label="Resumo do pedido">
      <header className="card-header">
        <h2>{summary.customerName}</h2>
        <OrderStatusBadge status={order.status} />
      </header>
      <p className="muted">
        Criado em {summary.createdAt} · {summary.units}
      </p>
      {summary.lines.length === 0 ? (
        <p className="muted">O pedido ainda não tem itens.</p>
      ) : (
        <table className="table">
          <thead>
            <tr>
              <th>Produto</th>
              <th className="numeric">Quantidade</th>
              <th className="numeric">Preço unitário</th>
              <th className="numeric">Subtotal</th>
            </tr>
          </thead>
          <tbody>
            {summary.lines.map((line) => (
              <ItemLine key={line.key} line={line} />
            ))}
          </tbody>
        </table>
      )}
      <p className="total">
        Total <strong>{summary.total}</strong>
      </p>
    </section>
  );
}
