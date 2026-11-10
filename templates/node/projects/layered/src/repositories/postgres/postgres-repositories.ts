import type { Repositories } from '../repositories.js';
import type { PostgresDatabase } from './postgres-database.js';
import { PostgresTaskRepository } from './postgres-task-repository.js';
// romanctl:imports

export function postgresRepositories(database: PostgresDatabase): Repositories {
  return {
    tasks: new PostgresTaskRepository(database),
    // romanctl:repositories
  };
}
