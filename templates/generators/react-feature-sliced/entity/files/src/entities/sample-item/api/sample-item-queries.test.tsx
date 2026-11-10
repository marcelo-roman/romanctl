import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';
import { server, withQueryClient } from '@tests/support';
import { useSampleItem, useSampleItems } from './sample-item-queries';

const wrapper = ({ children }: { children: ReactNode }) => withQueryClient(children);

const dto = { id: 'a-1', name: 'Ana', createdAt: '2026-01-02T10:00:00Z' };

describe('sample item queries', () => {
  it('reads the list and maps every sample item', async () => {
    server.use(http.get('*/api/sample-items', () => HttpResponse.json([dto])));

    const { result } = renderHook(() => useSampleItems(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([{ id: 'a-1', name: 'Ana', createdAt: new Date(dto.createdAt) }]);
  });

  it('exposes the problem detail when the sample item does not exist', async () => {
    server.use(
      http.get('*/api/sample-items/:id', () =>
        HttpResponse.json({ status: 404, title: 'Não encontrado', detail: 'não encontrado' }, { status: 404 }),
      ),
    );

    const { result } = renderHook(() => useSampleItem('missing'), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toMatchObject({ status: 404, detail: 'não encontrado' });
  });
});
