import { Pool, type PoolClient, type QueryResultRow } from 'pg';
import { schema } from './schema.js';

export class PostgresDatabase {
  private constructor(private readonly pool: Pool) {}

  static async open(connectionString: string): Promise<PostgresDatabase> {
    const database = new PostgresDatabase(new Pool({ connectionString }));
    try {
      await database.createSchema();
      return database;
    } catch (error) {
      await database.close();
      throw error;
    }
  }

  async query<Row extends QueryResultRow>(sql: string, params: unknown[] = []): Promise<Row[]> {
    const result = await this.pool.query<Row>(sql, params);
    return result.rows;
  }

  async execute(sql: string, params: unknown[] = []): Promise<number> {
    const result = await this.pool.query(sql, params);
    return result.rowCount ?? 0;
  }

  async transaction(work: (client: PoolClient) => Promise<void>): Promise<void> {
    const client = await this.pool.connect();
    try {
      await client.query('BEGIN');
      await work(client);
      await client.query('COMMIT');
    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally {
      client.release();
    }
  }

  async close(): Promise<void> {
    await this.pool.end();
  }

  private createSchema(): Promise<void> {
    return this.transaction(async (client) => {
      await client.query(`SELECT pg_advisory_xact_lock(hashtext('schema'))`);
      await client.query(schema);
    });
  }
}
