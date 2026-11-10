import { Inject } from '@nestjs/common';
import { Command, CommandHandler, type ICommandHandler } from '@nestjs/cqrs';
import { SampleItemId } from '../../../domain/sample-items/sample-item-id.js';
import { SampleItemRepository } from '../../../domain/sample-items/sample-item-repository.js';
import { NotFoundError } from '../../common/not-found-error.js';

export class RenameSampleItem extends Command<void> {
  constructor(
    readonly id: string,
    readonly name: string,
  ) {
    super();
  }
}

@CommandHandler(RenameSampleItem)
export class RenameSampleItemHandler implements ICommandHandler<RenameSampleItem, void> {
  constructor(@Inject(SampleItemRepository) private readonly repository: SampleItemRepository) {}

  async execute(command: RenameSampleItem): Promise<void> {
    const sampleItem = await this.repository.findById(SampleItemId.parse(command.id));
    if (!sampleItem) {
      throw new NotFoundError(`SampleItem ${command.id} não encontrado.`);
    }
    sampleItem.rename(command.name);
    await this.repository.save(sampleItem);
  }
}
