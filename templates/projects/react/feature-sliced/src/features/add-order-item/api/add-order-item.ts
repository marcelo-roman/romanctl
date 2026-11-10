import { postJson } from '@/shared/api';

export interface AddItemInput {
  productName: string;
  quantity: number;
  unitPrice: number;
  currency?: string;
}

export function addOrderItem(orderId: string, input: AddItemInput): Promise<void> {
  return postJson<void>(`/orders/${encodeURIComponent(orderId)}/items`, input);
}
