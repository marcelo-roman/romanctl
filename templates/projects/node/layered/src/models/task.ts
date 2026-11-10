export const TITLE_MAX_LENGTH = 200;

export interface Task {
  id: string;
  title: string;
  done: boolean;
  dueDate: string | null;
  createdAt: Date;
}

export interface TaskFilter {
  done?: boolean;
  limit: number;
}
