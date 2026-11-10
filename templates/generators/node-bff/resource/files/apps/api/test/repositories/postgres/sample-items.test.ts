import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { SampleItem } from '../../../src/domain/sample-item.js';
import type { PostgresDatabase } from '../../../src/repositories/postgres/postgres-database.js';
import { PostgresSampleItemRepository } from '../../../src/repositories/postgres/postgres-sample-item-repository.js';
import { databaseUrl, openDatabase, suiteTitle } from './database.js';

describe.skipIf(!databaseUrl)(suiteTitle('SampleItems in PostgreSQL'), () => {
  let db: PostgresDatabase;
  let repository: PostgresSampleItemRepository;

  beforeAll(async () => {
    db = await openDatabase();
    repository = new PostgresSampleItemRepository(db);
  });

  afterAll(() => db?.close());

  it('saves, renames, lists and removes', async () => {
    const sampleItem = SampleItem.create('Primeiro', new Date(Date.now() + 60_000));
    await repository.save(sampleItem);
    expect(await repository.findById(sampleItem.id)).toEqual(sampleItem);

    sampleItem.rename('Segundo');
    await repository.save(sampleItem);
    expect((await repository.findById(sampleItem.id))?.name).toBe('Segundo');
    expect((await repository.latest(1)).map((item) => item.id)).toEqual([sampleItem.id]);

    expect(await repository.remove(sampleItem.id)).toBe(true);
    expect(await repository.findById(sampleItem.id)).toBeUndefined();
  });
});
