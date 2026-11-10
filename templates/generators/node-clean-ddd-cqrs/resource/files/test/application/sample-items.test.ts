import { CommandBus, CqrsModule, QueryBus } from '@nestjs/cqrs';
import { Test } from '@nestjs/testing';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { Clock } from '../../src/application/common/clock.js';
import { NotFoundError } from '../../src/application/common/not-found-error.js';
import { CreateSampleItem } from '../../src/application/sample-items/commands/create-sample-item.js';
import { DeleteSampleItem } from '../../src/application/sample-items/commands/delete-sample-item.js';
import { RenameSampleItem } from '../../src/application/sample-items/commands/rename-sample-item.js';
import { GetSampleItemById } from '../../src/application/sample-items/queries/get-sample-item-by-id.js';
import { ListSampleItems } from '../../src/application/sample-items/queries/list-sample-items.js';
import { sampleItemHandlers } from '../../src/application/sample-items/sample-item-handlers.js';
import { SampleItemReadStore } from '../../src/application/sample-items/sample-item-read-store.js';
import { SampleItemRepository } from '../../src/domain/sample-items/sample-item-repository.js';
import { DomainError } from '../../src/domain/shared/domain-error.js';
import { InMemorySampleItemReadStore } from '../../src/infrastructure/persistence/memory/in-memory-sample-item-read-store.js';
import { InMemorySampleItemRepository } from '../../src/infrastructure/persistence/memory/in-memory-sample-item-repository.js';
import { InMemorySampleItemStore } from '../../src/infrastructure/persistence/memory/in-memory-sample-item-store.js';

class FixedClock extends Clock {
  now(): Date {
    return new Date('2026-01-01T00:00:00Z');
  }
}

describe('SampleItem use cases', () => {
  let commands: CommandBus;
  let queries: QueryBus;
  let close: () => Promise<void>;

  beforeEach(async () => {
    const module = await Test.createTestingModule({
      imports: [CqrsModule.forRoot()],
      providers: [
        ...sampleItemHandlers,
        InMemorySampleItemStore,
        { provide: SampleItemRepository, useClass: InMemorySampleItemRepository },
        { provide: SampleItemReadStore, useClass: InMemorySampleItemReadStore },
        { provide: Clock, useClass: FixedClock },
      ],
    }).compile();
    const app = await module.init();
    commands = app.get(CommandBus);
    queries = app.get(QueryBus);
    close = () => app.close();
  });

  afterEach(() => close());

  it('creates, renames, lists and deletes', async () => {
    const id = await commands.execute(new CreateSampleItem('Primeiro'));
    await commands.execute(new RenameSampleItem(id, 'Segundo'));

    expect(await queries.execute(new GetSampleItemById(id))).toEqual({ id, name: 'Segundo', createdAt: '2026-01-01T00:00:00.000Z' });
    expect(await queries.execute(new ListSampleItems())).toHaveLength(1);

    await commands.execute(new DeleteSampleItem(id));
    await expect(queries.execute(new GetSampleItemById(id))).rejects.toThrow(NotFoundError);
    await expect(commands.execute(new DeleteSampleItem(id))).rejects.toThrow(NotFoundError);
  });

  it('refuses an empty name', async () => {
    await expect(commands.execute(new CreateSampleItem(' '))).rejects.toThrow(DomainError);
  });
});
