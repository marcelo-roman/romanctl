import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { SampleItemsPage } from '../src/pages/SampleItemsPage';
import { mockFetch } from './fetch-mock';

const first = { id: 'r1', name: 'Primeiro', createdOn: '10/03/2026' };
const second = { id: 'r2', name: 'Segundo', createdOn: '11/03/2026' };

describe('SampleItemsPage', () => {
  it('lists, creates and removes through the BFF', async () => {
    const calls = mockFetch((call) => {
      if (call.method === 'POST') {
        return { body: { sampleItems: [second, first] } };
      }
      if (call.method === 'DELETE') {
        return { body: { sampleItems: [second] } };
      }
      return { body: { sampleItems: [first] } };
    });
    render(<SampleItemsPage />);
    expect(await screen.findByText('Primeiro')).toBeTruthy();

    await userEvent.type(screen.getByLabelText('Nome'), 'Segundo');
    await userEvent.click(screen.getByRole('button', { name: 'Adicionar' }));
    expect(await screen.findByText('Segundo')).toBeTruthy();
    expect(calls.at(-1)).toEqual({ method: 'POST', url: '/actions/sample-items', body: { name: 'Segundo' } });

    await userEvent.click(screen.getByRole('button', { name: 'Remover Primeiro' }));
    expect(calls.at(-1)).toMatchObject({ method: 'DELETE', url: '/actions/sample-items/r1' });
    expect(screen.queryByText('Primeiro')).toBeNull();
  });

  it('shows the message the BFF sends', async () => {
    mockFetch((call) => (call.method === 'POST' ? { status: 422, body: { message: 'O nome é obrigatório.' } } : { body: { sampleItems: [] } }));
    render(<SampleItemsPage />);
    await screen.findByText('Nenhum registro ainda.');

    await userEvent.click(screen.getByRole('button', { name: 'Adicionar' }));

    expect((await screen.findByRole('alert')).textContent).toBe('O nome é obrigatório.');
  });
});
