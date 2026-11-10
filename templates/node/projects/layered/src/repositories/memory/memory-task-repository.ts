import type { Task, TaskFilter } from '../../models/task.js';
import type { TaskRepository } from '../task-repository.js';

export class MemoryTaskRepository implements TaskRepository {
  private readonly tasks = new Map<string, Task>();

  async add(task: Task): Promise<void> {
    this.tasks.set(task.id, { ...task });
  }

  async findById(id: string): Promise<Task | undefined> {
    const task = this.tasks.get(id);
    return task ? { ...task } : undefined;
  }

  async list(filter: TaskFilter): Promise<Task[]> {
    return [...this.tasks.values()]
      .filter((task) => filter.done === undefined || task.done === filter.done)
      .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
      .slice(0, filter.limit)
      .map((task) => ({ ...task }));
  }

  async update(task: Task): Promise<void> {
    this.tasks.set(task.id, { ...task });
  }

  async remove(id: string): Promise<boolean> {
    return this.tasks.delete(id);
  }
}
