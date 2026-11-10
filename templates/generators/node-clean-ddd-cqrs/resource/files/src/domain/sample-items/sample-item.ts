import { AggregateRoot } from '../shared/aggregate-root.js';
import { DomainError } from '../shared/domain-error.js';
import { SampleItemId } from './sample-item-id.js';

export class SampleItem extends AggregateRoot {
  private constructor(
    readonly id: SampleItemId,
    private currentName: string,
    readonly createdAt: Date,
  ) {
    super();
  }

  static create(name: string, now: Date): SampleItem {
    const sampleItem = new SampleItem(SampleItemId.create(), '', now);
    sampleItem.rename(name);
    return sampleItem;
  }

  static restore(id: SampleItemId, name: string, createdAt: Date): SampleItem {
    return new SampleItem(id, name, createdAt);
  }

  get name(): string {
    return this.currentName;
  }

  rename(name: string): void {
    if (name.trim() === '') {
      throw new DomainError('O nome é obrigatório.');
    }
    this.currentName = name.trim();
  }
}
