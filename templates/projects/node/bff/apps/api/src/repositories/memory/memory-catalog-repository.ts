import type { CatalogProduct } from '../../domain/catalog-product.js';
import type { CatalogRepository } from '../catalog-repository.js';

export class MemoryCatalogRepository implements CatalogRepository {
  private readonly products = new Map<string, CatalogProduct>();

  async add(product: CatalogProduct): Promise<void> {
    this.products.set(product.id, product);
  }

  async findById(id: string): Promise<CatalogProduct | undefined> {
    return this.products.get(id);
  }

  async list(): Promise<CatalogProduct[]> {
    return [...this.products.values()].sort((a, b) => a.name.localeCompare(b.name, 'pt-BR'));
  }
}
