import { buildApp } from './app.js';
import { PostgresDatabase } from './repositories/postgres/postgres-database.js';
import { postgresRepositories } from './repositories/postgres/postgres-repositories.js';

const DEFAULT_DATABASE_URL = 'postgres://postgres:postgres@127.0.0.1:5432/sample_app';

const port = Number(process.env.PORT) || 3001;
const database = await PostgresDatabase.open(process.env.DATABASE_URL || DEFAULT_DATABASE_URL);
const app = buildApp(postgresRepositories(database), { logger: true });
app.addHook('onClose', () => database.close());

for (const signal of ['SIGINT', 'SIGTERM'] as const) {
  process.once(signal, () => void app.close());
}

await app.listen({ port, host: '0.0.0.0' });
