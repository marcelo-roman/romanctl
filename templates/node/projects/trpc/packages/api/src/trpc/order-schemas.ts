import { z } from 'zod';

const orderId = z.uuid({ error: 'Invalid order identifier.' });

export const createOrderInput = z.object({
  customerName: z
    .string({ error: 'Enter the customer name.' })
    .trim()
    .min(1, { error: 'Enter the customer name.' })
    .max(200, { error: 'The customer name has at most 200 characters.' }),
});

export const addItemInput = z.object({
  orderId,
  productName: z
    .string({ error: 'Enter the product name.' })
    .trim()
    .min(1, { error: 'Enter the product name.' })
    .max(200, { error: 'The product name has at most 200 characters.' }),
  quantity: z
    .number({ error: 'Enter the quantity.' })
    .int({ error: 'The quantity must be a whole number.' })
    .positive({ error: 'The quantity must be greater than zero.' }),
  unitPrice: z.number({ error: 'Enter the unit price.' }).nonnegative({ error: 'The price cannot be negative.' }),
  currency: z
    .string()
    .trim()
    .regex(/^[A-Za-z]{3}$/, { error: 'The currency must have 3 letters, like USD.' })
    .optional(),
});

export const orderIdInput = z.object({ orderId });

export const listOrdersInput = z.object({ limit: z.number().int().optional() }).optional();
