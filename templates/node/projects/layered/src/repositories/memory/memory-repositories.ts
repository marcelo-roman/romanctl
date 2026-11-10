import type { Repositories } from '../repositories.js';
import { MemoryTaskRepository } from './memory-task-repository.js';
// romanctl:imports

export function memoryRepositories(): Repositories {
  return {
    tasks: new MemoryTaskRepository(),
    // romanctl:repositories
  };
}
