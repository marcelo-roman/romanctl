import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { SampleItemsPage } from '../src/pages/SampleItemsPage';
import { mockFetch } from './fetch-mock';

const first = { id: 'r1', name: 'First', createdOn: '3/10/26' };
const second = { id: 'r2', name: 'Second', createdOn: '3/11/26' };

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
    expect(await screen.findByText('First')).toBeTruthy();

    await userEvent.type(screen.getByLabelText('Name'), 'Second');
    await userEvent.click(screen.getByRole('button', { name: 'Add' }));
    expect(await screen.findByText('Second')).toBeTruthy();
    expect(calls.at(-1)).toEqual({ method: 'POST', url: '/actions/sample-items', body: { name: 'Second' } });

    await userEvent.click(screen.getByRole('button', { name: 'Remove First' }));
    expect(calls.at(-1)).toMatchObject({ method: 'DELETE', url: '/actions/sample-items/r1' });
    expect(screen.queryByText('First')).toBeNull();
  });

  it('shows the message the BFF sends', async () => {
    mockFetch((call) => (call.method === 'POST' ? { status: 422, body: { message: 'The name is required.' } } : { body: { sampleItems: [] } }));
    render(<SampleItemsPage />);
    await screen.findByText('No items yet.');

    await userEvent.click(screen.getByRole('button', { name: 'Add' }));

    expect((await screen.findByRole('alert')).textContent).toBe('The name is required.');
  });
});
