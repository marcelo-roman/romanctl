import { Inject, Injectable } from '@nestjs/common';
import { SampleItemReadStore } from '../../../application/sample-items/sample-item-read-store.js';
import type { SampleItemView } from '../../../application/sample-items/sample-item-view.js';
import { PostgresDatabase } from './postgres-database.js';

interface SampleItemRow {
  id: string;
  name: string;
  created_at: Date;
}

@Injectable()
export class PostgresSampleItemReadStore extends SampleItemReadStore {
  constructor(@Inject(PostgresDatabase) private readonly db: PostgresDatabase) {
    super();
  }

  async findById(id: string): Promise<SampleItemView | undefined> {
    const [row] = await this.db.query<SampleItemRow>('SELECT id, name, created_at FROM sample_items WHERE id = $1', [id]);
    return row && viewOf(row);
  }

  async list(limit: number): Promise<SampleItemView[]> {
    const rows = await this.db.query<SampleItemRow>(
      'SELECT id, name, created_at FROM sample_items ORDER BY created_at DESC LIMIT $1',
      [limit],
    );
    return rows.map(viewOf);
  }
}

function viewOf(row: SampleItemRow): SampleItemView {
  return { id: row.id, name: row.name, createdAt: row.created_at.toISOString() };
}
