export const OrderStatus = {
  Draft: 'Draft',
  Placed: 'Placed',
} as const;

export type OrderStatus = (typeof OrderStatus)[keyof typeof OrderStatus];
