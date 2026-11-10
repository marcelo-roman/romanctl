import { CatalogProduct } from '../../domain/catalog-product.js';
import { Money } from '../../domain/money.js';
import type { CatalogRepository } from '../catalog-repository.js';
import type { PostgresDatabase } from './postgres-database.js';

interface CatalogProductRow {
  id: string;
  name: string;
  price: string;
  currency: string;
  created_at: Date;
}

const columns = 'id, name, price, currency, created_at';

function toProduct(row: CatalogProductRow): CatalogProduct {
  return CatalogProduct.restore(row.id, row.name, Money.fromCents(Number(row.price), row.currency), row.created_at);
}

export class PostgresCatalogRepository implements CatalogRepository {
  constructor(private readonly database: PostgresDatabase) {}

  async add(product: CatalogProduct): Promise<void> {
    await this.database.execute('INSERT INTO catalog_products (id, name, price, currency, created_at) VALUES ($1, $2, $3, $4, $5)', [
      product.id,
      product.name,
      product.price.cents,
      product.price.currency,
      product.createdAt,
    ]);
  }

  async findById(id: string): Promise<CatalogProduct | undefined> {
    const [row] = await this.database.query<CatalogProductRow>(`SELECT ${columns} FROM catalog_products WHERE id = $1`, [id]);
    return row ? toProduct(row) : undefined;
  }

  async list(): Promise<CatalogProduct[]> {
    const rows = await this.database.query<CatalogProductRow>(`SELECT ${columns} FROM catalog_products ORDER BY name, id`);
    return rows.map(toProduct);
  }
}
