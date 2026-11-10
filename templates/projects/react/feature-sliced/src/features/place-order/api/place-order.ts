import { postJson } from '@/shared/api';

export function placeOrder(orderId: string): Promise<void> {
  return postJson<void>(`/orders/${encodeURIComponent(orderId)}/place`);
}
