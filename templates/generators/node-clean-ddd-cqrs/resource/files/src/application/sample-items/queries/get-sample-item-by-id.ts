import { Inject } from '@nestjs/common';
import { type IQueryHandler, Query, QueryHandler } from '@nestjs/cqrs';
import { SampleItemId } from '../../../domain/sample-items/sample-item-id.js';
import { NotFoundError } from '../../common/not-found-error.js';
import { SampleItemReadStore } from '../sample-item-read-store.js';
import type { SampleItemView } from '../sample-item-view.js';

export class GetSampleItemById extends Query<SampleItemView> {
  constructor(readonly id: string) {
    super();
  }
}

@QueryHandler(GetSampleItemById)
export class GetSampleItemByIdHandler implements IQueryHandler<GetSampleItemById, SampleItemView> {
  constructor(@Inject(SampleItemReadStore) private readonly store: SampleItemReadStore) {}

  async execute(query: GetSampleItemById): Promise<SampleItemView> {
    const sampleItem = await this.store.findById(SampleItemId.parse(query.id).value);
    if (!sampleItem) {
      throw new NotFoundError(`SampleItem ${query.id} não encontrado.`);
    }
    return sampleItem;
  }
}
