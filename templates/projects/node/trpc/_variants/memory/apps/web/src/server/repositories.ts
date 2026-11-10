import { memoryRepositories, type Repositories } from '@sample-app/db';

const cache = globalThis as typeof globalThis & { sampleAppRepositories?: Repositories };

export async function repositories(): Promise<Repositories> {
  cache.sampleAppRepositories ??= memoryRepositories();
  return cache.sampleAppRepositories;
}
