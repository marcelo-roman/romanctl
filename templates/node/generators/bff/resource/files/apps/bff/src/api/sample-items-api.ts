import type { ApiHttp } from './api-http.js';

export interface ApiSampleItem {
  id: string;
  name: string;
  createdAt: string;
}

const notFound = 'Registro não encontrado.';

export class SampleItemsApi {
  constructor(private readonly http: ApiHttp) {}

  list(limit: number): Promise<ApiSampleItem[]> {
    return this.http.get(`/sample-items?limit=${limit}`);
  }

  create(name: string): Promise<ApiSampleItem> {
    return this.http.post('/sample-items', { body: { name } });
  }

  rename(id: string, name: string): Promise<ApiSampleItem> {
    return this.http.put(`/sample-items/${encodeURIComponent(id)}`, { body: { name }, notFound });
  }

  delete(id: string): Promise<void> {
    return this.http.delete(`/sample-items/${encodeURIComponent(id)}`, { notFound });
  }
}
