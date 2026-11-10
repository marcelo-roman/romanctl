import { addItemInput, createOrderInput, listOrdersInput, orderIdInput } from './order-schemas';
import { procedure, router } from './trpc';

export const ordersRouter = router({
  create: procedure.input(createOrderInput).mutation(({ ctx, input }) => ctx.orders.create(input.customerName)),
  addItem: procedure.input(addItemInput).mutation(({ ctx, input }) => ctx.orders.addItem(input)),
  place: procedure.input(orderIdInput).mutation(({ ctx, input }) => ctx.orders.place(input.orderId)),
  byId: procedure.input(orderIdInput).query(({ ctx, input }) => ctx.orders.byId(input.orderId)),
  list: procedure.input(listOrdersInput).query(({ ctx, input }) => ctx.orders.list(input?.limit)),
});
