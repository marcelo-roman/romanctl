import { ordersBff } from '../bff/orders';
import { ErrorMessage } from '../components/ErrorMessage';
import { NewOrderForm } from '../components/NewOrderForm';
import { OrdersTable } from '../components/OrdersTable';
import { useAction } from '../hooks/useAction';
import { useScreen } from '../hooks/useScreen';
import { navigate } from '../router';

export function OrdersPage() {
  const screen = useScreen('orders', ordersBff.list);
  const action = useAction();

  async function create(customerName: string) {
    const created = await action.run(() => ordersBff.create(customerName));
    if (created) {
      navigate(`/orders/${created.id}`);
    }
  }

  return (
    <section>
      <h1>Pedidos</h1>
      <NewOrderForm pending={action.pending} onCreate={create} />
      <ErrorMessage message={action.error ?? screen.error} />
      {screen.loading && !screen.data && <p>Carregando...</p>}
      {screen.data && <OrdersTable orders={screen.data.orders} />}
    </section>
  );
}
