import { existsSync } from 'node:fs';
import { buildApp } from './app.js';

const DEFAULT_API_URL = 'http://localhost:3302';

const port = Number(process.env.BFF_PORT) || 3301;
const webDist = process.env.WEB_DIST;
const app = buildApp({
  apiUrl: process.env.API_URL || DEFAULT_API_URL,
  webDist: webDist && existsSync(webDist) ? webDist : undefined,
  logger: true,
});

for (const signal of ['SIGINT', 'SIGTERM'] as const) {
  process.once(signal, () => void app.close());
}

await app.listen({ port, host: '0.0.0.0' });
