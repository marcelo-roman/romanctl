import type { CatalogApi } from '../api/catalog-api.js';
import type { OrdersApi } from '../api/orders-api.js';
import type { ApiCatalogProduct, ApiOrder } from '../api/api-types.js';
import { formatMoney } from './money.js';
import { statusLabel } from './status-label.js';

const LIST_LIMIT = 50;

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

function itemCount(order: ApiOrder): number {
  return order.items.reduce((count, item) => count + item.quantity, 0);
}

function orderRow(order: ApiOrder): OrderRow {
  return {
    id: order.id,
    customerName: order.customerName,
    status: statusLabel(order.status),
    itemCount: itemCount(order),
    total: formatMoney(order.total, order.currency),
  };
}

function suggestion(product: ApiCatalogProduct): ProductSuggestion {
  return { id: product.id, name: product.name, price: formatMoney(product.price, product.currency) };
}

export class OrdersScreen {
  constructor(
    private readonly orders: OrdersApi,
    private readonly catalog: CatalogApi,
  ) {}

  async list(): Promise<OrdersListScreen> {
    const orders = await this.orders.list(LIST_LIMIT);
    return { orders: orders.map(orderRow) };
  }

  async detail(id: string): Promise<OrderDetailScreen> {
    const [order, products] = await Promise.all([this.orders.get(id), this.catalog.list()]);
    return this.compose(order, products);
  }

  async afterChange(order: ApiOrder): Promise<OrderDetailScreen> {
    return this.compose(order, order.canAddItems ? await this.catalog.list() : []);
  }

  private compose(order: ApiOrder, products: ApiCatalogProduct[]): OrderDetailScreen {
    return {
      order: {
        id: order.id,
        customerName: order.customerName,
        status: statusLabel(order.status),
        total: formatMoney(order.total, order.currency),
        canAddItems: order.canAddItems,
        canPlace: order.canPlace,
        items: order.items.map((item) => ({
          id: item.id,
          productName: item.productName,
          quantity: item.quantity,
          unitPrice: formatMoney(item.unitPrice, order.currency),
          subtotal: formatMoney(item.subtotal, order.currency),
        })),
      },
      suggestions: order.canAddItems ? products.map(suggestion) : [],
    };
  }
}
