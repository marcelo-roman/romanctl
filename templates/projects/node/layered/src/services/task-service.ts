import { randomUUID } from 'node:crypto';
import { TITLE_MAX_LENGTH, type Task } from '../models/task.js';
import type { TaskRepository } from '../repositories/task-repository.js';
import { type Clock, isoDate } from './clock.js';
import { ConflictError, NotFoundError, RuleViolationError, ValidationError } from './errors.js';
import { listLimit } from './list-limit.js';

export interface CreateTaskInput {
  title: string;
  dueDate?: string | null;
}

export interface UpdateTaskInput {
  title?: string;
  done?: boolean;
  dueDate?: string | null;
}

export interface ListTasksInput {
  done?: boolean;
  limit?: number;
}

function requiredTitle(title: string): string {
  const trimmed = title.trim();
  if (trimmed === '') {
    throw new ValidationError('O título é obrigatório.');
  }
  if (trimmed.length > TITLE_MAX_LENGTH) {
    throw new ValidationError(`O título deve ter no máximo ${TITLE_MAX_LENGTH} caracteres.`);
  }
  return trimmed;
}

export class TaskService {
  constructor(
    private readonly tasks: TaskRepository,
    private readonly clock: Clock,
  ) {}

  async create(input: CreateTaskInput): Promise<Task> {
    const title = requiredTitle(input.title);
    const now = this.clock();
    const dueDate = input.dueDate ?? null;
    if (dueDate !== null && dueDate < isoDate(now)) {
      throw new RuleViolationError('A data de entrega não pode estar no passado.');
    }
    const task: Task = { id: randomUUID(), title, done: false, dueDate, createdAt: now };
    await this.tasks.add(task);
    return task;
  }

  list(input: ListTasksInput): Promise<Task[]> {
    return this.tasks.list({ done: input.done, limit: listLimit(input.limit) });
  }

  async get(id: string): Promise<Task> {
    const task = await this.tasks.findById(id);
    if (!task) {
      throw new NotFoundError(`Tarefa ${id} não encontrada.`);
    }
    return task;
  }

  async update(id: string, input: UpdateTaskInput): Promise<Task> {
    const task = await this.get(id);
    if (input.done === true && task.done) {
      throw new ConflictError('A tarefa já está concluída.');
    }
    const updated: Task = {
      ...task,
      title: input.title === undefined ? task.title : requiredTitle(input.title),
      done: input.done ?? task.done,
      dueDate: input.dueDate === undefined ? task.dueDate : input.dueDate,
    };
    await this.tasks.update(updated);
    return updated;
  }

  async delete(id: string): Promise<void> {
    const removed = await this.tasks.remove(id);
    if (!removed) {
      throw new NotFoundError(`Tarefa ${id} não encontrada.`);
    }
  }
}
