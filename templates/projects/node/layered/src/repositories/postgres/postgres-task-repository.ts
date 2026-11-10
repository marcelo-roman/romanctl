import type { Task, TaskFilter } from '../../models/task.js';
import type { TaskRepository } from '../task-repository.js';
import type { PostgresDatabase } from './postgres-database.js';

interface TaskRow {
  id: string;
  title: string;
  done: boolean;
  due_date: string | null;
  created_at: Date;
}

const columns = 'id, title, done, due_date::text AS due_date, created_at';

function toTask(row: TaskRow): Task {
  return { id: row.id, title: row.title, done: row.done, dueDate: row.due_date, createdAt: row.created_at };
}

export class PostgresTaskRepository implements TaskRepository {
  constructor(private readonly database: PostgresDatabase) {}

  async add(task: Task): Promise<void> {
    await this.database.execute('INSERT INTO tasks (id, title, done, due_date, created_at) VALUES ($1, $2, $3, $4, $5)', [
      task.id,
      task.title,
      task.done,
      task.dueDate,
      task.createdAt,
    ]);
  }

  async findById(id: string): Promise<Task | undefined> {
    const [row] = await this.database.query<TaskRow>(`SELECT ${columns} FROM tasks WHERE id = $1`, [id]);
    return row ? toTask(row) : undefined;
  }

  async list(filter: TaskFilter): Promise<Task[]> {
    const rows = await this.database.query<TaskRow>(
      `SELECT ${columns} FROM tasks WHERE ($1::boolean IS NULL OR done = $1) ORDER BY created_at DESC, id LIMIT $2`,
      [filter.done ?? null, filter.limit],
    );
    return rows.map(toTask);
  }

  async update(task: Task): Promise<void> {
    await this.database.execute('UPDATE tasks SET title = $2, done = $3, due_date = $4 WHERE id = $1', [task.id, task.title, task.done, task.dueDate]);
  }

  async remove(id: string): Promise<boolean> {
    return (await this.database.execute('DELETE FROM tasks WHERE id = $1', [id])) > 0;
  }
}
