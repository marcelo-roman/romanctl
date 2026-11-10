import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { SampleItem } from '../../../src/domain/sample-items/sample-item.js';
import type { PostgresDatabase } from '../../../src/infrastructure/persistence/postgres/postgres-database.js';
import { PostgresSampleItemReadStore } from '../../../src/infrastructure/persistence/postgres/postgres-sample-item-read-store.js';
import { PostgresSampleItemRepository } from '../../../src/infrastructure/persistence/postgres/postgres-sample-item-repository.js';
import { databaseUrl, openDatabase, suiteTitle } from './database.js';

describe.skipIf(!databaseUrl)(suiteTitle('SampleItems in PostgreSQL'), () => {
  let db: PostgresDatabase;
  let repository: PostgresSampleItemRepository;
  let readStore: PostgresSampleItemReadStore;

  beforeAll(async () => {
    db = await openDatabase();
    repository = new PostgresSampleItemRepository(db);
    readStore = new PostgresSampleItemReadStore(db);
  });

  afterAll(() => db?.onModuleDestroy());

  it('saves, renames, reads and removes', async () => {
    const sampleItem = SampleItem.create('Primeiro', new Date(Date.now() + 60_000));
    await repository.save(sampleItem);
    sampleItem.rename('Segundo');
    await repository.save(sampleItem);

    expect((await repository.findById(sampleItem.id))?.name).toBe('Segundo');
    expect(await readStore.findById(sampleItem.id.value)).toMatchObject({ id: sampleItem.id.value, name: 'Segundo' });
    expect((await readStore.list(1)).map((view) => view.id)).toEqual([sampleItem.id.value]);

    await repository.remove(sampleItem);
    expect(await repository.findById(sampleItem.id)).toBeUndefined();
  });
});
