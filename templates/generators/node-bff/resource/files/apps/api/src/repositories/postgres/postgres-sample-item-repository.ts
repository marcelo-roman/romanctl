import { SampleItem } from '../../domain/sample-item.js';
import type { SampleItemRepository } from '../sample-item-repository.js';
import type { PostgresDatabase } from './postgres-database.js';

interface SampleItemRow {
  id: string;
  name: string;
  created_at: Date;
}

function toSampleItem(row: SampleItemRow): SampleItem {
  return SampleItem.restore(row.id, row.name, row.created_at);
}

export class PostgresSampleItemRepository implements SampleItemRepository {
  constructor(private readonly database: PostgresDatabase) {}

  async save(sampleItem: SampleItem): Promise<void> {
    await this.database.execute(
      `INSERT INTO sample_items (id, name, created_at) VALUES ($1, $2, $3)
       ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`,
      [sampleItem.id, sampleItem.name, sampleItem.createdAt],
    );
  }

  async findById(id: string): Promise<SampleItem | undefined> {
    const [row] = await this.database.query<SampleItemRow>('SELECT id, name, created_at FROM sample_items WHERE id = $1', [id]);
    return row ? toSampleItem(row) : undefined;
  }

  async latest(limit: number): Promise<SampleItem[]> {
    const rows = await this.database.query<SampleItemRow>('SELECT id, name, created_at FROM sample_items ORDER BY created_at DESC, id LIMIT $1', [limit]);
    return rows.map(toSampleItem);
  }

  async remove(id: string): Promise<boolean> {
    return (await this.database.execute('DELETE FROM sample_items WHERE id = $1', [id])) > 0;
  }
}
