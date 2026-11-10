import type { Task, TaskFilter } from '../models/task.js';

export interface TaskRepository {
  add(task: Task): Promise<void>;
  findById(id: string): Promise<Task | undefined>;
  list(filter: TaskFilter): Promise<Task[]>;
  update(task: Task): Promise<void>;
  remove(id: string): Promise<boolean>;
}
