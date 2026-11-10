import { PostgresDatabase } from '../../../src/repositories/postgres/postgres-database.js';

export const databaseUrl = process.env.DATABASE_URL;

export function suiteTitle(name: string): string {
  return databaseUrl ? name : `${name} (skipped: DATABASE_URL not set, romanctl up --no-run creates the database and the .env)`;
}

export function openDatabase(): Promise<PostgresDatabase> {
  return PostgresDatabase.open(databaseUrl ?? '');
}
