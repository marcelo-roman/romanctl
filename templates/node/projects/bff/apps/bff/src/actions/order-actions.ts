import type { CatalogApi } from '../api/catalog-api.js';
import type { OrdersApi } from '../api/orders-api.js';
import type { OrderDetailScreen, OrdersScreen } from '../screens/orders-screen.js';

export interface CreateOrderAction {
  customerName: string;
}

export interface AddItemAction {
  productId: string;
  quantity: number;
}

export interface CreatedOrder {
  id: string;
}

export class OrderActions {
  constructor(
    private readonly orders: OrdersApi,
    private readonly catalog: CatalogApi,
    private readonly screen: OrdersScreen,
  ) {}

  async create(action: CreateOrderAction): Promise<CreatedOrder> {
    const order = await this.orders.create(action.customerName);
    return { id: order.id };
  }

  async addItem(orderId: string, action: AddItemAction): Promise<OrderDetailScreen> {
    const product = await this.catalog.get(action.productId);
    const order = await this.orders.addItem(orderId, {
      productName: product.name,
      quantity: action.quantity,
      unitPrice: product.price,
      currency: product.currency,
    });
    return this.screen.afterChange(order);
  }

  async place(orderId: string): Promise<OrderDetailScreen> {
    return this.screen.afterChange(await this.orders.place(orderId));
  }
}
