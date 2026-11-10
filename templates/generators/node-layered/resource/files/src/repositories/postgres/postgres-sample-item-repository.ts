import type { SampleItem } from '../../models/sample-item.js';
import type { SampleItemRepository } from '../sample-item-repository.js';
import type { PostgresDatabase } from './postgres-database.js';

interface SampleItemRow {
  id: string;
  name: string;
  created_at: Date;
}

function toSampleItem(row: SampleItemRow): SampleItem {
  return { id: row.id, name: row.name, createdAt: row.created_at };
}

export class PostgresSampleItemRepository implements SampleItemRepository {
  constructor(private readonly database: PostgresDatabase) {}

  async add(sampleItem: SampleItem): Promise<void> {
    await this.database.execute('INSERT INTO sample_items (id, name, created_at) VALUES ($1, $2, $3)', [
      sampleItem.id,
      sampleItem.name,
      sampleItem.createdAt,
    ]);
  }

  async findById(id: string): Promise<SampleItem | undefined> {
    const [row] = await this.database.query<SampleItemRow>('SELECT id, name, created_at FROM sample_items WHERE id = $1', [id]);
    return row ? toSampleItem(row) : undefined;
  }

  async list(limit: number): Promise<SampleItem[]> {
    const rows = await this.database.query<SampleItemRow>('SELECT id, name, created_at FROM sample_items ORDER BY created_at DESC, id LIMIT $1', [limit]);
    return rows.map(toSampleItem);
  }

  async update(sampleItem: SampleItem): Promise<void> {
    await this.database.execute('UPDATE sample_items SET name = $2 WHERE id = $1', [sampleItem.id, sampleItem.name]);
  }

  async remove(id: string): Promise<boolean> {
    return (await this.database.execute('DELETE FROM sample_items WHERE id = $1', [id])) > 0;
  }
}
