import type { OnModuleDestroy, OnModuleInit } from '@nestjs/common';
import { Pool, type PoolClient, type QueryResultRow } from 'pg';
import { schema } from './schema.js';

export class PostgresDatabase implements OnModuleInit, OnModuleDestroy {
  private readonly pool: Pool;

  constructor(connectionString: string) {
    this.pool = new Pool({ connectionString });
  }

  onModuleInit(): Promise<void> {
    return this.transaction(async (client) => {
      await client.query(`SELECT pg_advisory_xact_lock(hashtext('schema'))`);
      await client.query(schema);
    });
  }

  async onModuleDestroy(): Promise<void> {
    await this.pool.end();
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
}
