import { postJson } from '@/shared/api';

interface CreatedOrder {
  id: string;
}

export async function createOrder(customerName: string): Promise<string> {
  const created = await postJson<CreatedOrder>('/orders', { customerName });
  return created.id;
}
