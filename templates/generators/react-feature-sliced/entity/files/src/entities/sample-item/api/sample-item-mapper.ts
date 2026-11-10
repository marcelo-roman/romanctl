import type { SampleItem } from '../model/sample-item';
import type { SampleItemDto } from './sample-item-dto';

export function sampleItemOf(dto: SampleItemDto): SampleItem {
  return { id: dto.id, name: dto.name, createdAt: new Date(dto.createdAt) };
}
