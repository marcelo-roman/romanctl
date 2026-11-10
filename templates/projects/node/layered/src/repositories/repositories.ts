import type { TaskRepository } from './task-repository.js';
// romanctl:imports

export interface Repositories {
  tasks: TaskRepository;
  // romanctl:repositories
}
