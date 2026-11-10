import { buildApp } from './app.js';
import { memoryRepositories } from './repositories/memory/memory-repositories.js';
import { CatalogService } from './services/catalog-service.js';
import { systemClock } from './services/clock.js';
import { starterCatalog } from './services/starter-catalog.js';

const port = Number(process.env.API_PORT) || 3302;
const repositories = memoryRepositories();
await new CatalogService(repositories.catalog, systemClock).seed(starterCatalog);
const app = buildApp(repositories, { logger: true });

for (const signal of ['SIGINT', 'SIGTERM'] as const) {
  process.once(signal, () => void app.close());
}

await app.listen({ port, host: '0.0.0.0' });
