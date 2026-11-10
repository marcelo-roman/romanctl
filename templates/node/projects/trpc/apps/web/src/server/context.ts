import { type Context, createContext } from '@sample-app/api';
import { repositories } from './repositories';

export async function createRequestContext(): Promise<Context> {
  return createContext(await repositories());
}
