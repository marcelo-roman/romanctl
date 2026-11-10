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
    const created = await app.inject({ method: 'POST', url: '/tasks', payload: { title: ' Comprar pão ', dueDate: '2026-03-12' } });
    expect(created.statusCode).toBe(201);
    const task = created.json();
    expect(created.headers.location).toBe(`/tasks/${task.id}`);
    expect(task).toEqual({ id: task.id, title: 'Comprar pão', done: false, dueDate: '2026-03-12', createdAt: now.toISOString() });

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
    const first = await createTask({ title: 'Um' });
    await createTask({ title: 'Dois' });
    await app.inject({ method: 'PATCH', url: `/tasks/${first.id}`, payload: { done: true } });

    const done = await app.inject({ method: 'GET', url: '/tasks?done=true' });
    expect(done.json().map((task: { title: string }) => task.title)).toEqual(['Um']);

    const limited = await app.inject({ method: 'GET', url: '/tasks?limit=1' });
    expect(limited.json()).toHaveLength(1);
  });

  it('answers 409 when completing a task twice', async () => {
    const task = await createTask({ title: 'Lavar louça' });
    await app.inject({ method: 'PATCH', url: `/tasks/${task.id}`, payload: { done: true } });

    const again = await app.inject({ method: 'PATCH', url: `/tasks/${task.id}`, payload: { done: true } });

    expect(again.statusCode).toBe(409);
    expect(again.json()).toEqual({ status: 409, title: 'Conflito', detail: 'A tarefa já está concluída.' });
  });

  it('answers 422 for a due date in the past', async () => {
    const response = await app.inject({ method: 'POST', url: '/tasks', payload: { title: 'Pagar conta', dueDate: '2026-03-09' } });

    expect(response.statusCode).toBe(422);
    expect(response.headers['content-type']).toContain('application/problem+json');
    expect(response.json()).toEqual({ status: 422, title: 'Regra de negócio violada', detail: 'A data de entrega não pode estar no passado.' });
  });

  it('answers 400 for invalid requests', async () => {
    const cases: [InjectOptions, string][] = [
      [{ method: 'POST', url: '/tasks', payload: {} }, 'O campo title é obrigatório.'],
      [{ method: 'POST', url: '/tasks', payload: { title: '   ' } }, 'O título é obrigatório.'],
      [{ method: 'POST', url: '/tasks', payload: { title: 'a'.repeat(201) } }, 'O campo title deve ter no máximo 200 caracteres.'],
      [{ method: 'POST', url: '/tasks', payload: { title: 'Ler', dueDate: '10/03/2026' } }, 'O campo dueDate deve estar no formato AAAA-MM-DD.'],
      [{ method: 'POST', url: '/tasks', headers: { 'content-type': 'application/json' }, payload: '{' }, 'O corpo não é um JSON válido.'],
      [{ method: 'GET', url: '/tasks?done=talvez' }, 'O campo done deve ser verdadeiro ou falso.'],
      [{ method: 'GET', url: '/tasks/123' }, 'O campo id deve estar no formato UUID.'],
      [{ method: 'PATCH', url: `/tasks/${missingId}`, payload: {} }, 'Informe ao menos um campo.'],
    ];
    for (const [request, detail] of cases) {
      const response = await app.inject(request);
      expect(response.statusCode, detail).toBe(400);
      expect(response.json()).toEqual({ status: 400, title: 'Requisição inválida', detail });
    }
  });

  it('answers 404 for unknown tasks', async () => {
    const response = await app.inject({ method: 'GET', url: `/tasks/${missingId}` });

    expect(response.statusCode).toBe(404);
    expect(response.json()).toMatchObject({ status: 404, title: 'Não encontrado' });
    expect((await app.inject({ method: 'DELETE', url: `/tasks/${missingId}` })).statusCode).toBe(404);
  });
});
