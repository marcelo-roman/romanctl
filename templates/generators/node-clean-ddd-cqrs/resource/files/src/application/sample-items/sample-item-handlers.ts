import { CreateSampleItemHandler } from './commands/create-sample-item.js';
import { DeleteSampleItemHandler } from './commands/delete-sample-item.js';
import { RenameSampleItemHandler } from './commands/rename-sample-item.js';
import { GetSampleItemByIdHandler } from './queries/get-sample-item-by-id.js';
import { ListSampleItemsHandler } from './queries/list-sample-items.js';

export const sampleItemHandlers = [
  CreateSampleItemHandler,
  RenameSampleItemHandler,
  DeleteSampleItemHandler,
  GetSampleItemByIdHandler,
  ListSampleItemsHandler,
];
