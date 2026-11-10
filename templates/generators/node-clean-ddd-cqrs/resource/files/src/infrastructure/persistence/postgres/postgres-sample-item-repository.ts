import { Inject, Injectable } from '@nestjs/common';
import { SampleItem } from '../../../domain/sample-items/sample-item.js';
import { SampleItemId } from '../../../domain/sample-items/sample-item-id.js';
import { SampleItemRepository } from '../../../domain/sample-items/sample-item-repository.js';
import { PostgresDatabase } from './postgres-database.js';

interface SampleItemRow {
  id: string;
  name: string;
  created_at: Date;
}

@Injectable()
export class PostgresSampleItemRepository extends SampleItemRepository {
  constructor(@Inject(PostgresDatabase) private readonly db: PostgresDatabase) {
    super();
  }

  async save(sampleItem: SampleItem): Promise<void> {
    await this.db.execute(
      `INSERT INTO sample_items (id, name, created_at) VALUES ($1, $2, $3)
       ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`,
      [sampleItem.id.value, sampleItem.name, sampleItem.createdAt],
    );
  }

  async findById(id: SampleItemId): Promise<SampleItem | undefined> {
    const [row] = await this.db.query<SampleItemRow>('SELECT id, name, created_at FROM sample_items WHERE id = $1', [id.value]);
    return row && SampleItem.restore(SampleItemId.parse(row.id), row.name, row.created_at);
  }

  async remove(sampleItem: SampleItem): Promise<void> {
    await this.db.execute('DELETE FROM sample_items WHERE id = $1', [sampleItem.id.value]);
  }
}
