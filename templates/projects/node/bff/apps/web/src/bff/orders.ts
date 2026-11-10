import { bffRequest } from './bff-client';

export interface OrderRow {
  id: string;
  customerName: string;
  status: string;
  itemCount: number;
  total: string;
}

export interface OrdersListScreen {
  orders: OrderRow[];
}

export interface OrderLine {
  id: string;
  productName: string;
  quantity: number;
  unitPrice: string;
  subtotal: string;
}

export interface ProductSuggestion {
  id: string;
  name: string;
  price: string;
}

export interface OrderDetailScreen {
  order: {
    id: string;
    customerName: string;
    status: string;
    total: string;
    canAddItems: boolean;
    canPlace: boolean;
    items: OrderLine[];
  };
  suggestions: ProductSuggestion[];
}

export interface CreatedOrder {
  id: string;
}

const orderPath = (id: string) => `/orders/${encodeURIComponent(id)}`;

export const ordersBff = {
  list: () => bffRequest<OrdersListScreen>('/screens/orders'),
  detail: (id: string) => bffRequest<OrderDetailScreen>(`/screens${orderPath(id)}`),
  create: (customerName: string) => bffRequest<CreatedOrder>('/actions/orders', { method: 'POST', body: { customerName } }),
  addItem: (id: string, productId: string, quantity: number) =>
    bffRequest<OrderDetailScreen>(`/actions${orderPath(id)}/items`, { method: 'POST', body: { productId, quantity } }),
  place: (id: string) => bffRequest<OrderDetailScreen>(`/actions${orderPath(id)}/place`, { method: 'POST' }),
};
