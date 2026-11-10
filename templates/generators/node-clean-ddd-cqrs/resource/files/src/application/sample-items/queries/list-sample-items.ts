import { Inject } from '@nestjs/common';
import { type IQueryHandler, Query, QueryHandler } from '@nestjs/cqrs';
import { listLimit } from '../../common/list-limit.js';
import { SampleItemReadStore } from '../sample-item-read-store.js';
import type { SampleItemView } from '../sample-item-view.js';

export class ListSampleItems extends Query<SampleItemView[]> {
  constructor(readonly limit?: number) {
    super();
  }
}

@QueryHandler(ListSampleItems)
export class ListSampleItemsHandler implements IQueryHandler<ListSampleItems, SampleItemView[]> {
  constructor(@Inject(SampleItemReadStore) private readonly store: SampleItemReadStore) {}

  execute(query: ListSampleItems): Promise<SampleItemView[]> {
    return this.store.list(listLimit(query.limit));
  }
}
