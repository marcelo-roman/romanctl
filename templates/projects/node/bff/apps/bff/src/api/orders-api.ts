import type { ApiHttp } from './api-http.js';
import type { ApiNewOrderItem, ApiOrder } from './api-types.js';

const notFound = 'Pedido não encontrado.';

export class OrdersApi {
  constructor(private readonly http: ApiHttp) {}

  list(limit: number): Promise<ApiOrder[]> {
    return this.http.get(`/orders?limit=${limit}`);
  }

  get(id: string): Promise<ApiOrder> {
    return this.http.get(`/orders/${encodeURIComponent(id)}`, { notFound });
  }

  create(customerName: string): Promise<ApiOrder> {
    return this.http.post('/orders', { body: { customerName } });
  }

  addItem(id: string, item: ApiNewOrderItem): Promise<ApiOrder> {
    return this.http.post(`/orders/${encodeURIComponent(id)}/items`, { body: item, notFound });
  }

  place(id: string): Promise<ApiOrder> {
    return this.http.post(`/orders/${encodeURIComponent(id)}/place`, { notFound });
  }
}
