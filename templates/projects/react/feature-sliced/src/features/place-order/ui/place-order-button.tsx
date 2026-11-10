import { Alert, Button } from '@/shared/ui';
import { usePlaceOrder } from '../model/use-place-order';

export function PlaceOrderButton({ orderId }: { orderId: string }) {
  const order = usePlaceOrder(orderId);

  return (
    <div className="actions">
      <Button onClick={order.place} disabled={order.isPending}>
        {order.isPending ? 'Realizando...' : 'Realizar pedido'}
      </Button>
      {order.error && <Alert tone="error">{order.error}</Alert>}
    </div>
  );
}
