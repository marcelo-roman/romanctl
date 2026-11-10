import { Inject } from '@nestjs/common';
import { Command, CommandHandler, type ICommandHandler } from '@nestjs/cqrs';
import { SampleItem } from '../../../domain/sample-items/sample-item.js';
import { SampleItemRepository } from '../../../domain/sample-items/sample-item-repository.js';
import { Clock } from '../../common/clock.js';

export class CreateSampleItem extends Command<string> {
  constructor(readonly name: string) {
    super();
  }
}

@CommandHandler(CreateSampleItem)
export class CreateSampleItemHandler implements ICommandHandler<CreateSampleItem, string> {
  constructor(
    @Inject(SampleItemRepository) private readonly repository: SampleItemRepository,
    @Inject(Clock) private readonly clock: Clock,
  ) {}

  async execute(command: CreateSampleItem): Promise<string> {
    const sampleItem = SampleItem.create(command.name, this.clock.now());
    await this.repository.save(sampleItem);
    return sampleItem.id.value;
  }
}
