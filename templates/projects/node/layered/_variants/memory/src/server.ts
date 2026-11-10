import { buildApp } from './app.js';
import { memoryRepositories } from './repositories/memory/memory-repositories.js';

const port = Number(process.env.PORT) || 3001;
const app = buildApp(memoryRepositories(), { logger: true });

for (const signal of ['SIGINT', 'SIGTERM'] as const) {
  process.once(signal, () => void app.close());
}

await app.listen({ port, host: '0.0.0.0' });
