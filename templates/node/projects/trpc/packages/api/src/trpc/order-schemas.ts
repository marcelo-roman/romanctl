import { z } from 'zod';

const orderId = z.uuid({ error: 'Identificador de pedido inválido.' });

export const createOrderInput = z.object({
  customerName: z
    .string({ error: 'Informe o nome do cliente.' })
    .trim()
    .min(1, { error: 'Informe o nome do cliente.' })
    .max(200, { error: 'O nome do cliente tem no máximo 200 caracteres.' }),
});

export const addItemInput = z.object({
  orderId,
  productName: z
    .string({ error: 'Informe o nome do produto.' })
    .trim()
    .min(1, { error: 'Informe o nome do produto.' })
    .max(200, { error: 'O nome do produto tem no máximo 200 caracteres.' }),
  quantity: z
    .number({ error: 'Informe a quantidade.' })
    .int({ error: 'A quantidade deve ser um número inteiro.' })
    .positive({ error: 'A quantidade deve ser maior que zero.' }),
  unitPrice: z.number({ error: 'Informe o preço unitário.' }).nonnegative({ error: 'O preço não pode ser negativo.' }),
  currency: z
    .string()
    .trim()
    .regex(/^[A-Za-z]{3}$/, { error: 'A moeda deve ter 3 letras, como BRL.' })
    .optional(),
});

export const orderIdInput = z.object({ orderId });

export const listOrdersInput = z.object({ limit: z.number().int().optional() }).optional();
