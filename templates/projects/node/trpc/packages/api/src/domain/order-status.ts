export const OrderStatus = {
  Draft: 'Draft',
  Placed: 'Placed',
} as const;

export type OrderStatus = (typeof OrderStatus)[keyof typeof OrderStatus];

export function parseOrderStatus(value: string): OrderStatus {
  const status = Object.values(OrderStatus).find((known) => known === value);
  if (!status) {
    throw new Error(`Unknown order status: ${value}`);
  }
  return status;
}
