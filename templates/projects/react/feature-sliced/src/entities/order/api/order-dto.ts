export interface OrderItemDto {
  productName: string;
  quantity: number;
  unitPrice: number;
  subtotal: number;
}

export interface OrderDto {
  id: string;
  customerName: string;
  status: string;
  createdAt: string;
  total: number;
  currency: string;
  items: OrderItemDto[];
}
