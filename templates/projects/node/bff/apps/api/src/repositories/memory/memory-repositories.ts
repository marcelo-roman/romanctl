import type { Repositories } from '../repositories.js';
import { MemoryCatalogRepository } from './memory-catalog-repository.js';
import { MemoryOrderRepository } from './memory-order-repository.js';
// romanctl:imports

export function memoryRepositories(): Repositories {
  return {
    orders: new MemoryOrderRepository(),
    catalog: new MemoryCatalogRepository(),
    // romanctl:repositories
  };
}
