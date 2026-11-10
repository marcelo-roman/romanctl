import { z } from 'zod';

const id = z.uuid({ error: 'Identificador inválido.' });
const name = z
  .string({ error: 'Informe o nome.' })
  .trim()
  .min(1, { error: 'Informe o nome.' })
  .max(200, { error: 'O nome tem no máximo 200 caracteres.' });

export const createSampleItemInput = z.object({ name });

export const renameSampleItemInput = z.object({ id, name });

export const sampleItemIdInput = z.object({ id });

export const listSampleItemsInput = z.object({ limit: z.number().int().optional() }).optional();
