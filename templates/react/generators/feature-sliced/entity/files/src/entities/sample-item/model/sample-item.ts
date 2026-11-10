export interface SampleItem {
  id: string;
  name: string;
  createdAt: Date;
}

const unnamed = 'Sem nome';

export function sampleItemLabel(sampleItem: SampleItem): string {
  return sampleItem.name.trim() || unnamed;
}

export function newestFirst(sampleItems: SampleItem[]): SampleItem[] {
  return [...sampleItems].sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
}
