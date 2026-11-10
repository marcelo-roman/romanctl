import Fastify, { type FastifyInstance } from 'fastify';
import type { Repositories } from './repositories/repositories.js';
import { healthRoutes } from './routes/health.js';
import { sendNotFound, sendProblem } from './routes/problem.js';
import { taskRoutes } from './routes/tasks.js';
import { type Clock, systemClock } from './services/clock.js';
import { TaskService } from './services/task-service.js';
// romanctl:imports

export interface AppOptions {
  logger?: boolean;
  clock?: Clock;
}

export function buildApp(repositories: Repositories, options: AppOptions = {}): FastifyInstance {
  const clock = options.clock ?? systemClock;
  const app = Fastify({ logger: options.logger ?? false });
  app.setErrorHandler(sendProblem);
  app.setNotFoundHandler(sendNotFound);
  app.register(healthRoutes);
  app.register(taskRoutes(new TaskService(repositories.tasks, clock)));
  // romanctl:routes
  return app;
}
