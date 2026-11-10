import { PostgresDatabase } from '../../../src/infrastructure/persistence/postgres/postgres-database.js';

export const databaseUrl = process.env.DATABASE_URL;

export function suiteTitle(name: string): string {
  return databaseUrl ? name : `${name} (skipped: DATABASE_URL not set, romanctl up --no-run creates the database and the .env)`;
}

export async function openDatabase(): Promise<PostgresDatabase> {
  const db = new PostgresDatabase(databaseUrl ?? '');
  await db.onModuleInit();
  return db;
}
