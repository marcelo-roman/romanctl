import { beforeEach, describe, expect, it } from 'vitest';
import { MemoryTaskRepository } from '../../src/repositories/memory/memory-task-repository.js';
import { ConflictError, NotFoundError, RuleViolationError, ValidationError } from '../../src/services/errors.js';
import { TaskService } from '../../src/services/task-service.js';

const now = new Date('2026-03-10T12:00:00Z');
const missingId = '00000000-0000-4000-8000-000000000000';

describe('TaskService', () => {
  let service: TaskService;

  beforeEach(() => {
    service = new TaskService(new MemoryTaskRepository(), () => now);
  });

  it('creates a task with a trimmed title', async () => {
    const task = await service.create({ title: '  Comprar pão  ', dueDate: '2026-03-10' });

    expect(task).toMatchObject({ title: 'Comprar pão', done: false, dueDate: '2026-03-10', createdAt: now });
    expect(await service.get(task.id)).toEqual(task);
  });

  it('requires a title', async () => {
    await expect(service.create({ title: '   ' })).rejects.toThrow(ValidationError);
    await expect(service.create({ title: 'a'.repeat(201) })).rejects.toThrow(ValidationError);
  });

  it('refuses a due date in the past on create', async () => {
    await expect(service.create({ title: 'Pagar conta', dueDate: '2026-03-09' })).rejects.toThrow(RuleViolationError);
  });

  it('updates title, done and due date', async () => {
    const task = await service.create({ title: 'Ler', dueDate: '2026-04-01' });

    const updated = await service.update(task.id, { title: ' Ler livro ', done: true, dueDate: null });

    expect(updated).toMatchObject({ id: task.id, title: 'Ler livro', done: true, dueDate: null });
    expect(await service.get(task.id)).toEqual(updated);
  });

  it('refuses to complete a task twice', async () => {
    const task = await service.create({ title: 'Lavar louça' });
    await service.update(task.id, { done: true });

    await expect(service.update(task.id, { done: true })).rejects.toThrow(ConflictError);
    expect(await service.update(task.id, { done: false })).toMatchObject({ done: false });
  });

  it('filters by done and limits the list', async () => {
    const first = await service.create({ title: 'Um' });
    await service.create({ title: 'Dois' });
    await service.update(first.id, { done: true });

    expect((await service.list({ done: true })).map((task) => task.title)).toEqual(['Um']);
    expect((await service.list({ done: false })).map((task) => task.title)).toEqual(['Dois']);
    expect(await service.list({ limit: 1 })).toHaveLength(1);
  });

  it('reports unknown tasks', async () => {
    await expect(service.get(missingId)).rejects.toThrow(NotFoundError);
    await expect(service.update(missingId, { done: true })).rejects.toThrow(NotFoundError);
    await expect(service.delete(missingId)).rejects.toThrow(NotFoundError);
  });

  it('deletes a task', async () => {
    const task = await service.create({ title: 'Varrer' });

    await service.delete(task.id);

    await expect(service.get(task.id)).rejects.toThrow(NotFoundError);
  });
});
