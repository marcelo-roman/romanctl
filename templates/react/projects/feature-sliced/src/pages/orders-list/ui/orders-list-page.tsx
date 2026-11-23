import { CreateOrderForm } from '@/features/create-order';
import { OrdersTable } from '@/widgets/orders-table';

export function OrdersListPage() {
  return (
    <main className="page">
      <h1>Orders</h1>
      <CreateOrderForm />
      <OrdersTable />
    </main>
  );
}
