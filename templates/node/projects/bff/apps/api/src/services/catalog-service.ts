import { CatalogProduct } from '../domain/catalog-product.js';
import { Money } from '../domain/money.js';
import type { CatalogRepository } from '../repositories/catalog-repository.js';
import type { Clock } from './clock.js';
import { NotFoundError } from './errors.js';

export interface AddCatalogProductInput {
  name: string;
  price: number;
  currency?: string;
}

export class CatalogService {
  constructor(
    private readonly catalog: CatalogRepository,
    private readonly clock: Clock,
  ) {}

  async add(input: AddCatalogProductInput): Promise<CatalogProduct> {
    const product = CatalogProduct.create(input.name, Money.fromAmount(input.price, input.currency), this.clock());
    await this.catalog.add(product);
    return product;
  }

  list(): Promise<CatalogProduct[]> {
    return this.catalog.list();
  }

  async get(id: string): Promise<CatalogProduct> {
    const product = await this.catalog.findById(id);
    if (!product) {
      throw new NotFoundError(`Produto ${id} não encontrado no catálogo.`);
    }
    return product;
  }

  async seed(products: AddCatalogProductInput[]): Promise<void> {
    if ((await this.catalog.list()).length > 0) {
      return;
    }
    for (const product of products) {
      await this.add(product);
    }
  }
}
