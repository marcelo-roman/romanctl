import { buildApp } from './app.js';
import { PostgresDatabase } from './repositories/postgres/postgres-database.js';
import { postgresRepositories } from './repositories/postgres/postgres-repositories.js';
import { CatalogService } from './services/catalog-service.js';
import { systemClock } from './services/clock.js';
import { starterCatalog } from './services/starter-catalog.js';

const DEFAULT_DATABASE_URL = 'postgres://postgres:postgres@127.0.0.1:5432/sample_app';

const port = Number(process.env.API_PORT) || 3302;
const database = await PostgresDatabase.open(process.env.DATABASE_URL || DEFAULT_DATABASE_URL);
const repositories = postgresRepositories(database);
await new CatalogService(repositories.catalog, systemClock).seed(starterCatalog);
const app = buildApp(repositories, { logger: true });
app.addHook('onClose', () => database.close());

for (const signal of ['SIGINT', 'SIGTERM'] as const) {
  process.once(signal, () => void app.close());
}

await app.listen({ port, host: '0.0.0.0' });
