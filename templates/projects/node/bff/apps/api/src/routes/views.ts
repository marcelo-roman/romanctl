import type { CatalogProduct } from '../domain/catalog-product.js';
import type { Order } from '../domain/order.js';

export interface OrderItemView {
  id: string;
  productName: string;
  quantity: number;
  unitPrice: number;
  subtotal: number;
}

export interface OrderView {
  id: string;
  customerName: string;
  status: string;
  createdAt: string;
  currency: string;
  total: number;
  canAddItems: boolean;
  canPlace: boolean;
  items: OrderItemView[];
}

export interface CatalogProductView {
  id: string;
  name: string;
  price: number;
  currency: string;
}

export function orderView(order: Order): OrderView {
  return {
    id: order.id,
    customerName: order.customerName,
    status: order.status,
    createdAt: order.createdAt.toISOString(),
    currency: order.currency,
    total: order.total.amount,
    canAddItems: order.canAddItems,
    canPlace: order.canPlace,
    items: order.items.map((item) => ({
      id: item.id,
      productName: item.productName,
      quantity: item.quantity,
      unitPrice: item.unitPrice.amount,
      subtotal: item.subtotal.amount,
    })),
  };
}

export function catalogProductView(product: CatalogProduct): CatalogProductView {
  return { id: product.id, name: product.name, price: product.price.amount, currency: product.price.currency };
}
