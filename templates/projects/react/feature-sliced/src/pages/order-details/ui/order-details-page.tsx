import { Link, useParams } from 'react-router';
import { isDraft, useOrder, type Order } from '@/entities/order';
import { AddOrderItemForm } from '@/features/add-order-item';
import { PlaceOrderButton } from '@/features/place-order';
import { errorMessage } from '@/shared/api';
import { paths } from '@/shared/config';
import { Alert } from '@/shared/ui';
import { OrderSummary } from '@/widgets/order-summary';

function DraftActions({ order }: { order: Order }) {
  return (
    <section className="card" aria-label="Itens">
      <h2>Adicionar item</h2>
      <AddOrderItemForm orderId={order.id} />
      <PlaceOrderButton orderId={order.id} />
    </section>
  );
}

function OrderContent({ id }: { id: string }) {
  const { data, error, isPending } = useOrder(id);

  if (isPending) {
    return <p className="muted">Carregando pedido...</p>;
  }
  if (error) {
    return <Alert tone="error">{errorMessage(error)}</Alert>;
  }
  return (
    <>
      <OrderSummary order={data} />
      {isDraft(data) && <DraftActions order={data} />}
    </>
  );
}

export function OrderDetailsPage() {
  const { id = '' } = useParams();

  return (
    <main className="page">
      <Link to={paths.orders}>← Pedidos</Link>
      <h1>Pedido</h1>
      <OrderContent id={id} />
    </main>
  );
}
