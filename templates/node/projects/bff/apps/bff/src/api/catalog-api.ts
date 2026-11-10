import type { ApiHttp } from './api-http.js';
import type { ApiCatalogProduct } from './api-types.js';

export class CatalogApi {
  constructor(private readonly http: ApiHttp) {}

  list(): Promise<ApiCatalogProduct[]> {
    return this.http.get('/catalog/products');
  }

  get(id: string): Promise<ApiCatalogProduct> {
    return this.http.get(`/catalog/products/${encodeURIComponent(id)}`, { notFound: 'Produto não encontrado no catálogo.' });
  }
}
