import { statusLabel, type OrderStatus } from '../model/order';

export function OrderStatusBadge({ status }: { status: OrderStatus }) {
  return <span className={`badge badge-${status.toLowerCase()}`}>{statusLabel(status)}</span>;
}
