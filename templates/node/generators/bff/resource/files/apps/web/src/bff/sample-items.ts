import { bffRequest } from './bff-client';

export interface SampleItemRow {
  id: string;
  name: string;
  createdOn: string;
}

export interface SampleItemsScreen {
  sampleItems: SampleItemRow[];
}

export const sampleItemsBff = {
  list: () => bffRequest<SampleItemsScreen>('/screens/sample-items'),
  create: (name: string) => bffRequest<SampleItemsScreen>('/actions/sample-items', { method: 'POST', body: { name } }),
  delete: (id: string) => bffRequest<SampleItemsScreen>(`/actions/sample-items/${encodeURIComponent(id)}`, { method: 'DELETE' }),
};
