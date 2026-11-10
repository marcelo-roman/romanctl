export interface OrderItemRecord {
  id: string;
  productName: string;
  quantity: number;
  unitPriceCents: number;
  currency: string;
}

export interface OrderRecord {
  id: string;
  customerName: string;
  status: string;
  createdAt: Date;
  items: OrderItemRecord[];
}
