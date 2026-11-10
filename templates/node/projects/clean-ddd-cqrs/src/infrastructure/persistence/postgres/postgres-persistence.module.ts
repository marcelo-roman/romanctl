import { Module, type Provider } from '@nestjs/common';
import { OrderReadStore } from '../../../application/orders/order-read-store.js';
import { OrderRepository } from '../../../domain/orders/order-repository.js';
import { PostgresDatabase } from './postgres-database.js';
import { PostgresOrderLoader } from './postgres-order-loader.js';
import { PostgresOrderReadStore } from './postgres-order-read-store.js';
import { PostgresOrderRepository } from './postgres-order-repository.js';
// romanctl:imports

const DEFAULT_DATABASE_URL = 'postgres://postgres:postgres@127.0.0.1:5432/sample_app';

const providers: Provider[] = [
  { provide: PostgresDatabase, useFactory: () => new PostgresDatabase(process.env.DATABASE_URL || DEFAULT_DATABASE_URL) },
  PostgresOrderLoader,
  { provide: OrderRepository, useClass: PostgresOrderRepository },
  { provide: OrderReadStore, useClass: PostgresOrderReadStore },
  // romanctl:providers
];

@Module({ providers, exports: providers })
export class PostgresPersistenceModule {}
