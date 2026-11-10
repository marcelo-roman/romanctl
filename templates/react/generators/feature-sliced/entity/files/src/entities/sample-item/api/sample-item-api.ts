import { getJson } from '@/shared/api';
import type { SampleItem } from '../model/sample-item';
import type { SampleItemDto } from './sample-item-dto';
import { sampleItemOf } from './sample-item-mapper';

export async function fetchSampleItem(id: string): Promise<SampleItem> {
  return sampleItemOf(await getJson<SampleItemDto>(`/sample-items/${encodeURIComponent(id)}`));
}

export async function fetchSampleItems(): Promise<SampleItem[]> {
  const dtos = await getJson<SampleItemDto[]>('/sample-items');
  return dtos.map(sampleItemOf);
}
