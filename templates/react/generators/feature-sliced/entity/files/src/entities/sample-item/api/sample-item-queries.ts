import { useQuery } from '@tanstack/react-query';
import { fetchSampleItem, fetchSampleItems } from './sample-item-api';

export const sampleItemKeys = {
  all: ['sample-items'] as const,
  list: () => [...sampleItemKeys.all, 'list'] as const,
  detail: (id: string) => [...sampleItemKeys.all, 'detail', id] as const,
};

export function useSampleItem(id: string) {
  return useQuery({ queryKey: sampleItemKeys.detail(id), queryFn: () => fetchSampleItem(id) });
}

export function useSampleItems() {
  return useQuery({ queryKey: sampleItemKeys.list(), queryFn: fetchSampleItems });
}
