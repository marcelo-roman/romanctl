import type { ApiSampleItem, SampleItemsApi } from '../api/sample-items-api.js';

const LIST_LIMIT = 50;

const dateFormat = new Intl.DateTimeFormat('pt-BR', { dateStyle: 'short', timeZone: 'UTC' });

export interface SampleItemRow {
  id: string;
  name: string;
  createdOn: string;
}

export interface SampleItemsScreen {
  sampleItems: SampleItemRow[];
}

function sampleItemRow(sampleItem: ApiSampleItem): SampleItemRow {
  return { id: sampleItem.id, name: sampleItem.name, createdOn: dateFormat.format(new Date(sampleItem.createdAt)) };
}

export class SampleItemsScreens {
  constructor(private readonly api: SampleItemsApi) {}

  async list(): Promise<SampleItemsScreen> {
    const sampleItems = await this.api.list(LIST_LIMIT);
    return { sampleItems: sampleItems.map(sampleItemRow) };
  }

  async create(name: string): Promise<SampleItemsScreen> {
    await this.api.create(name);
    return this.list();
  }

  async rename(id: string, name: string): Promise<SampleItemsScreen> {
    await this.api.rename(id, name);
    return this.list();
  }

  async delete(id: string): Promise<SampleItemsScreen> {
    await this.api.delete(id);
    return this.list();
  }
}
