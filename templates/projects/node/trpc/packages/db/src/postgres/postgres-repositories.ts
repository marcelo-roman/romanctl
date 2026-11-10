import type { Repositories } from '../repositories';
import type { PostgresDatabase } from './postgres-database';
import { PostgresOrderRepository } from './postgres-order-repository';
// romanctl:imports

export function postgresRepositories(database: PostgresDatabase): Repositories {
  return {
    orders: new PostgresOrderRepository(database),
    // romanctl:repositories
  };
}
