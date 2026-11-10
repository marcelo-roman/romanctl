import { CreateOrderForm } from '../components/orders/create-order-form';
import { OrderList } from '../components/orders/order-list';

export default function HomePage() {
  return (
    <>
      <h1>Pedidos</h1>
      <CreateOrderForm />
      <OrderList />
    </>
  );
}
