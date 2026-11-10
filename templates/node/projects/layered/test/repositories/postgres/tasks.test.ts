import { randomUUID } from 'node:crypto';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import type { Task } from '../../../src/models/task.js';
import type { PostgresDatabase } from '../../../src/repositories/postgres/postgres-database.js';
import { PostgresTaskRepository } from '../../../src/repositories/postgres/postgres-task-repository.js';
import { databaseUrl, openDatabase, suiteTitle } from './database.js';

describe.skipIf(!databaseUrl)(suiteTitle('tasks in PostgreSQL'), () => {
  let db: PostgresDatabase;
  let repository: PostgresTaskRepository;

  beforeAll(async () => {
    db = await openDatabase();
    repository = new PostgresTaskRepository(db);
  });

  afterAll(() => db?.close());

  it('saves, updates, lists and removes a task', async () => {
    const task: Task = { id: randomUUID(), title: 'Comprar pão', done: false, dueDate: '2099-12-31', createdAt: new Date(Date.now() + 60_000) };
    await repository.add(task);
    expect(await repository.findById(task.id)).toEqual(task);

    const done: Task = { ...task, title: 'Comprar pão integral', done: true, dueDate: null };
    await repository.update(done);
    expect(await repository.findById(task.id)).toEqual(done);

    expect((await repository.list({ limit: 1 })).map((t) => t.id)).toEqual([task.id]);
    expect((await repository.list({ done: true, limit: 1 })).map((t) => t.id)).toEqual([task.id]);
    expect((await repository.list({ done: false, limit: 200 })).map((t) => t.id)).not.toContain(task.id);

    expect(await repository.remove(task.id)).toBe(true);
    expect(await repository.remove(task.id)).toBe(false);
    expect(await repository.findById(task.id)).toBeUndefined();
  });

  it('creates the schema again without failing', async () => {
    const again = await openDatabase();
    await again.close();
  });
});
