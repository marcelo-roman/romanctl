import type { Repositories } from '../repositories.js';
import { PostgresCatalogRepository } from './postgres-catalog-repository.js';
import type { PostgresDatabase } from './postgres-database.js';
import { PostgresOrderRepository } from './postgres-order-repository.js';
// romanctl:imports

export function postgresRepositories(database: PostgresDatabase): Repositories {
  return {
    orders: new PostgresOrderRepository(database),
    catalog: new PostgresCatalogRepository(database),
    // romanctl:repositories
  };
}
