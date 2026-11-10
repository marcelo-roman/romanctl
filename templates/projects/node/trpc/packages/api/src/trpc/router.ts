import { ordersRouter } from './orders-router';
import { createCallerFactory, router } from './trpc';
// romanctl:imports

export const appRouter = router({
  orders: ordersRouter,
  // romanctl:routers
});

export type AppRouter = typeof appRouter;

export const createCaller = createCallerFactory(appRouter);
