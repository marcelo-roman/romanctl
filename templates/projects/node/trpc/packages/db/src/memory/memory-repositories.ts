import type { Repositories } from '../repositories';
import { MemoryOrderRepository } from './memory-order-repository';
// romanctl:imports

export function memoryRepositories(): Repositories {
  return {
    orders: new MemoryOrderRepository(),
    // romanctl:repositories
  };
}
