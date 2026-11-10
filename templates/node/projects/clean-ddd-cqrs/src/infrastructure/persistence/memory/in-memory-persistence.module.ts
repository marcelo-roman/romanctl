import { Module, type Provider } from '@nestjs/common';
import { OrderReadStore } from '../../../application/orders/order-read-store.js';
import { OrderRepository } from '../../../domain/orders/order-repository.js';
import { InMemoryOrderReadStore } from './in-memory-order-read-store.js';
import { InMemoryOrderRepository } from './in-memory-order-repository.js';
import { InMemoryOrderStore } from './in-memory-order-store.js';
// romanctl:imports

const providers: Provider[] = [
  InMemoryOrderStore,
  { provide: OrderRepository, useClass: InMemoryOrderRepository },
  { provide: OrderReadStore, useClass: InMemoryOrderReadStore },
  // romanctl:providers
];

@Module({ providers, exports: providers })
export class InMemoryPersistenceModule {}
