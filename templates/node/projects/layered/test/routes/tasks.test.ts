import type { FastifyInstance, InjectOptions } from 'fastify';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { now, startApp } from './app.js';

const missingId = '00000000-0000-4000-8000-000000000000';

describe('tasks API', () => {
  let app: FastifyInstance;

  beforeEach(async () => {
    app = await startApp();
  });

  afterEach(() => app.close());

  async function createTask(payload: Record<string, unknown>): Promise<{ id: string }> {
    const response = await app.inject({ method: 'POST', url: '/tasks', payload });
    expect(response.statusCode).toBe(201);
    return response.json();
  }

  it('creates, reads, updates and deletes a task', async () => {
    const created = await app.inject({ method: 'POST', url: '/tasks', payload: { title: ' Buy bread ', dueDate: '2026-03-12' } });
    expect(created.statusCode).toBe(201);
    const task = created.json();
    expect(created.headers.location).toBe(`/tasks/${task.id}`);
    expect(task).toEqual({ id: task.id, title: 'Buy bread', done: false, dueDate: '2026-03-12', createdAt: now.toISOString() });

    const read = await app.inject({ method: 'GET', url: `/tasks/${task.id}` });
    expect(read.json()).toEqual(task);

    const updated = await app.inject({ method: 'PATCH', url: `/tasks/${task.id}`, payload: { done: true, dueDate: null } });
    expect(updated.statusCode).toBe(200);
    expect(updated.json()).toMatchObject({ done: true, dueDate: null });

    const deleted = await app.inject({ method: 'DELETE', url: `/tasks/${task.id}` });
    expect(deleted.statusCode).toBe(204);
    expect((await app.inject({ method: 'GET', url: `/tasks/${task.id}` })).statusCode).toBe(404);
  });

  it('lists tasks filtered by done with a limit', async () => {
    const first = await createTask({ title: 'One' });
    await createTask({ title: 'Two' });
    await app.inject({ method: 'PATCH', url: `/tasks/${first.id}`, payload: { done: true } });

    const done = await app.inject({ method: 'GET', url: '/tasks?done=true' });
    expect(done.json().map((task: { title: string }) => task.title)).toEqual(['One']);

    const limited = await app.inject({ method: 'GET', url: '/tasks?limit=1' });
    expect(limited.json()).toHaveLength(1);
  });

  it('answers 409 when completing a task twice', async () => {
    const task = await createTask({ title: 'Wash the dishes' });
    await app.inject({ method: 'PATCH', url: `/tasks/${task.id}`, payload: { done: true } });

    const again = await app.inject({ method: 'PATCH', url: `/tasks/${task.id}`, payload: { done: true } });

    expect(again.statusCode).toBe(409);
    expect(again.json()).toEqual({ status: 409, title: 'Conflict', detail: 'The task is already done.' });
  });

  it('answers 422 for a due date in the past', async () => {
    const response = await app.inject({ method: 'POST', url: '/tasks', payload: { title: 'Pay the bill', dueDate: '2026-03-09' } });

    expect(response.statusCode).toBe(422);
    expect(response.headers['content-type']).toContain('application/problem+json');
    expect(response.json()).toEqual({ status: 422, title: 'Business rule violated', detail: 'The due date cannot be in the past.' });
  });

  it('answers 400 for invalid requests', async () => {
    const cases: [InjectOptions, string][] = [
      [{ method: 'POST', url: '/tasks', payload: {} }, 'The field title is required.'],
      [{ method: 'POST', url: '/tasks', payload: { title: '   ' } }, 'Title is required.'],
      [{ method: 'POST', url: '/tasks', payload: { title: 'a'.repeat(201) } }, 'The field title must be at most 200 characters.'],
      [{ method: 'POST', url: '/tasks', payload: { title: 'Read', dueDate: '10/03/2026' } }, 'The field dueDate must be in the format YYYY-MM-DD.'],
      [{ method: 'POST', url: '/tasks', headers: { 'content-type': 'application/json' }, payload: '{' }, 'The body is not valid JSON.'],
      [{ method: 'GET', url: '/tasks?done=maybe' }, 'The field done must be true or false.'],
      [{ method: 'GET', url: '/tasks/123' }, 'The field id must be in the format UUID.'],
      [{ method: 'PATCH', url: `/tasks/${missingId}`, payload: {} }, 'Provide at least one field.'],
    ];
    for (const [request, detail] of cases) {
      const response = await app.inject(request);
      expect(response.statusCode, detail).toBe(400);
      expect(response.json()).toEqual({ status: 400, title: 'Invalid request', detail });
    }
  });

  it('answers 404 for unknown tasks', async () => {
    const response = await app.inject({ method: 'GET', url: `/tasks/${missingId}` });

    expect(response.statusCode).toBe(404);
    expect(response.json()).toMatchObject({ status: 404, title: 'Not found' });
    expect((await app.inject({ method: 'DELETE', url: `/tasks/${missingId}` })).statusCode).toBe(404);
  });
});
