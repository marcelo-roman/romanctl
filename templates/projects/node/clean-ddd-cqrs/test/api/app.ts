import type { INestApplication } from '@nestjs/common';
import { Test } from '@nestjs/testing';
import { AppModule } from '../../src/api/app.module.js';
import { InMemoryPersistenceModule } from '../../src/infrastructure/persistence/memory/in-memory-persistence.module.js';

export async function startApp(): Promise<INestApplication> {
  const module = await Test.createTestingModule({ imports: [AppModule.with(InMemoryPersistenceModule)] }).compile();
  const app = module.createNestApplication({ logger: false });
  return app.init();
}
