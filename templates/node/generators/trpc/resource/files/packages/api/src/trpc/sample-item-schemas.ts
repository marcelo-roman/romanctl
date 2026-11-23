import { z } from 'zod';

const id = z.uuid({ error: 'Invalid identifier.' });
const name = z
  .string({ error: 'Enter the name.' })
  .trim()
  .min(1, { error: 'Enter the name.' })
  .max(200, { error: 'The name has at most 200 characters.' });

export const createSampleItemInput = z.object({ name });

export const renameSampleItemInput = z.object({ id, name });

export const sampleItemIdInput = z.object({ id });

export const listSampleItemsInput = z.object({ limit: z.number().int().optional() }).optional();
