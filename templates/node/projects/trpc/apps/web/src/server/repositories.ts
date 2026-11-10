import type { Repositories } from '@sample-app/db';
import { PostgresDatabase, postgresRepositories } from '@sample-app/db/postgres';

const DEFAULT_DATABASE_URL = 'postgres://postgres:postgres@127.0.0.1:5432/sample_app';

const cache = globalThis as typeof globalThis & { sampleAppRepositories?: Promise<Repositories> };

function open(): Promise<Repositories> {
  return PostgresDatabase.open(process.env.DATABASE_URL || DEFAULT_DATABASE_URL)
    .then(postgresRepositories)
    .catch((error: unknown) => {
      cache.sampleAppRepositories = undefined;
      throw error;
    });
}

export function repositories(): Promise<Repositories> {
  cache.sampleAppRepositories ??= open();
  return cache.sampleAppRepositories;
}
