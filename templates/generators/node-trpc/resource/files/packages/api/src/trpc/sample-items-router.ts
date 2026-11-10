import { createSampleItemInput, listSampleItemsInput, renameSampleItemInput, sampleItemIdInput } from './sample-item-schemas';
import { procedure, router } from './trpc';

export const sampleItemsRouter = router({
  create: procedure.input(createSampleItemInput).mutation(({ ctx, input }) => ctx.sampleItems.create(input.name)),
  rename: procedure.input(renameSampleItemInput).mutation(({ ctx, input }) => ctx.sampleItems.rename(input.id, input.name)),
  remove: procedure.input(sampleItemIdInput).mutation(({ ctx, input }) => ctx.sampleItems.remove(input.id)),
  byId: procedure.input(sampleItemIdInput).query(({ ctx, input }) => ctx.sampleItems.byId(input.id)),
  list: procedure.input(listSampleItemsInput).query(({ ctx, input }) => ctx.sampleItems.list(input?.limit)),
});
