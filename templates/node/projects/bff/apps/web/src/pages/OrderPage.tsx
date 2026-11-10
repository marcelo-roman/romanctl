import { type OrderDetailScreen, ordersBff } from '../bff/orders';
import { AddItemForm } from '../components/AddItemForm';
import { ErrorMessage } from '../components/ErrorMessage';
import { OrderItemsTable } from '../components/OrderItemsTable';
import { useAction } from '../hooks/useAction';
import { useScreen } from '../hooks/useScreen';
import { Link } from '../router';

interface OrderPageProps {
  id: string;
}

export function OrderPage({ id }: OrderPageProps) {
  const screen = useScreen(`order:${id}`, () => ordersBff.detail(id));
  const action = useAction();

  async function change(work: () => Promise<OrderDetailScreen>) {
    const next = await action.run(work);
    if (next) {
      screen.replace(next);
    }
  }

  const order = screen.data?.order;
  return (
    <section>
      <Link to="/orders">Voltar para os pedidos</Link>
      <ErrorMessage message={action.error ?? screen.error} />
      {screen.loading && !order && <p>Carregando...</p>}
      {order && (
        <>
          <h1>Pedido de {order.customerName}</h1>
          <p>Situação: {order.status}</p>
          <OrderItemsTable items={order.items} total={order.total} />
          {order.canAddItems && (
            <AddItemForm
              suggestions={screen.data?.suggestions ?? []}
              pending={action.pending}
              onAdd={(productId, quantity) => change(() => ordersBff.addItem(id, productId, quantity))}
            />
          )}
          {order.canPlace && (
            <button type="button" disabled={action.pending} onClick={() => change(() => ordersBff.place(id))}>
              Realizar pedido
            </button>
          )}
        </>
      )}
    </section>
  );
}
