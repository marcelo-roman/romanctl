import { initTRPC } from '@trpc/server';
import type { Context } from '../context';
import { toTRPCError, validationIssues } from './error-mapping';

const t = initTRPC.context<Context>().create({
  errorFormatter({ shape, error }) {
    const issues = validationIssues(error);
    return { ...shape, message: issues[0] ?? shape.message, data: { ...shape.data, issues } };
  },
});

const mapErrors = t.middleware(async ({ next }) => {
  const result = await next();
  if (result.ok) {
    return result;
  }
  throw toTRPCError(result.error);
});

export const router = t.router;
export const procedure = t.procedure.use(mapErrors);
export const createCallerFactory = t.createCallerFactory;
