import type { SampleItemRecord } from '../sample-items/sample-item-record';
import type { SampleItemRepository } from '../sample-items/sample-item-repository';
import type { PostgresDatabase } from './postgres-database';

interface SampleItemRow {
  id: string;
  name: string;
  created_at: Date;
}

const selectSampleItems = 'SELECT id, name, created_at FROM sample_items';

function toRecord(row: SampleItemRow): SampleItemRecord {
  return { id: row.id, name: row.name, createdAt: row.created_at };
}

export class PostgresSampleItemRepository implements SampleItemRepository {
  constructor(private readonly database: PostgresDatabase) {}

  async save(sampleItem: SampleItemRecord): Promise<void> {
    await this.database.execute(
      `INSERT INTO sample_items (id, name, created_at) VALUES ($1, $2, $3)
       ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`,
      [sampleItem.id, sampleItem.name, sampleItem.createdAt],
    );
  }

  async findById(id: string): Promise<SampleItemRecord | undefined> {
    const [row] = await this.database.query<SampleItemRow>(`${selectSampleItems} WHERE id = $1`, [id]);
    return row && toRecord(row);
  }

  async latest(limit: number): Promise<SampleItemRecord[]> {
    const rows = await this.database.query<SampleItemRow>(`${selectSampleItems} ORDER BY created_at DESC LIMIT $1`, [limit]);
    return rows.map(toRecord);
  }

  async remove(id: string): Promise<boolean> {
    return (await this.database.execute('DELETE FROM sample_items WHERE id = $1', [id])) > 0;
  }
}
