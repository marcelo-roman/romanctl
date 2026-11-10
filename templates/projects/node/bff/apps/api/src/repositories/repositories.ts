import type { CatalogRepository } from './catalog-repository.js';
import type { OrderRepository } from './order-repository.js';
// romanctl:imports

export interface Repositories {
  orders: OrderRepository;
  catalog: CatalogRepository;
  // romanctl:repositories
}
