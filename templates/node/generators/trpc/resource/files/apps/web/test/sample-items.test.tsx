import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { SampleItemList } from '../src/components/sample-items/sample-item-list';
import { renderWithApi, testApi } from './api';

afterEach(cleanup);

function create(name: string) {
  fireEvent.change(screen.getByLabelText('Name'), { target: { value: name } });
  fireEvent.click(screen.getByRole('button', { name: 'Create' }));
}

describe('SampleItemList', () => {
  it('creates and removes, showing validation messages in English', async () => {
    renderWithApi(<SampleItemList />, testApi());
    expect(await screen.findByText('No records yet.')).toBeTruthy();

    create('   ');
    expect(await screen.findByText('Enter the name.')).toBeTruthy();

    create('First');
    expect(await screen.findByText('First')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Remove First' }));
    await waitFor(() => expect(screen.queryByText('First')).toBeNull());
  });
});
