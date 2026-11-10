import { Body, Controller, Delete, Get, HttpCode, HttpStatus, Inject, Param, Post, Put, Query, Res } from '@nestjs/common';
import { CommandBus, QueryBus } from '@nestjs/cqrs';
import { CreateSampleItem } from '../application/sample-items/commands/create-sample-item.js';
import { DeleteSampleItem } from '../application/sample-items/commands/delete-sample-item.js';
import { RenameSampleItem } from '../application/sample-items/commands/rename-sample-item.js';
import { GetSampleItemById } from '../application/sample-items/queries/get-sample-item-by-id.js';
import { ListSampleItems } from '../application/sample-items/queries/list-sample-items.js';
import type { SampleItemView } from '../application/sample-items/sample-item-view.js';
import type { Location } from './location.js';
import { RequestBody } from './request-body.js';

@Controller('sample-items')
export class SampleItemsController {
  constructor(
    @Inject(CommandBus) private readonly commands: CommandBus,
    @Inject(QueryBus) private readonly queries: QueryBus,
  ) {}

  @Post()
  async create(@Body() body: unknown, @Res({ passthrough: true }) response: Location): Promise<{ id: string }> {
    const id = await this.commands.execute(new CreateSampleItem(RequestBody.from(body).text('name')));
    response.location(`/sample-items/${id}`);
    return { id };
  }

  @Get()
  list(@Query('limit') limit?: string): Promise<SampleItemView[]> {
    return this.queries.execute(new ListSampleItems(limit ? Number(limit) : undefined));
  }

  @Get(':id')
  get(@Param('id') id: string): Promise<SampleItemView> {
    return this.queries.execute(new GetSampleItemById(id));
  }

  @Put(':id')
  @HttpCode(HttpStatus.NO_CONTENT)
  async rename(@Param('id') id: string, @Body() body: unknown): Promise<void> {
    await this.commands.execute(new RenameSampleItem(id, RequestBody.from(body).text('name')));
  }

  @Delete(':id')
  @HttpCode(HttpStatus.NO_CONTENT)
  async delete(@Param('id') id: string): Promise<void> {
    await this.commands.execute(new DeleteSampleItem(id));
  }
}
