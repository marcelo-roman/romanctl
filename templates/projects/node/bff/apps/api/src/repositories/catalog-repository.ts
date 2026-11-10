import type { CatalogProduct } from '../domain/catalog-product.js';

export interface CatalogRepository {
  add(product: CatalogProduct): Promise<void>;
  findById(id: string): Promise<CatalogProduct | undefined>;
  list(): Promise<CatalogProduct[]>;
}
