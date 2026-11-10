import type { FastifyPluginAsync } from 'fastify';
import type { CreateTaskInput, ListTasksInput, TaskService, UpdateTaskInput } from '../services/task-service.js';
import { createTaskSchema, deleteTaskSchema, getTaskSchema, listTasksSchema, updateTaskSchema } from './task-schemas.js';

interface IdParams {
  id: string;
}

export function taskRoutes(service: TaskService): FastifyPluginAsync {
  return async (app) => {
    app.post<{ Body: CreateTaskInput }>('/tasks', { schema: createTaskSchema }, async (request, reply) => {
      const task = await service.create(request.body);
      return reply.status(201).header('location', `/tasks/${task.id}`).send(task);
    });

    app.get<{ Querystring: ListTasksInput }>('/tasks', { schema: listTasksSchema }, (request) => service.list(request.query));

    app.get<{ Params: IdParams }>('/tasks/:id', { schema: getTaskSchema }, (request) => service.get(request.params.id));

    app.patch<{ Params: IdParams; Body: UpdateTaskInput }>('/tasks/:id', { schema: updateTaskSchema }, (request) =>
      service.update(request.params.id, request.body),
    );

    app.delete<{ Params: IdParams }>('/tasks/:id', { schema: deleteTaskSchema }, async (request, reply) => {
      await service.delete(request.params.id);
      return reply.status(204).send();
    });
  };
}
