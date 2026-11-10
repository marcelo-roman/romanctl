import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { SampleItemList } from '../src/components/sample-items/sample-item-list';
import { renderWithApi, testApi } from './api';

afterEach(cleanup);

function create(name: string) {
  fireEvent.change(screen.getByLabelText('Nome'), { target: { value: name } });
  fireEvent.click(screen.getByRole('button', { name: 'Criar' }));
}

describe('SampleItemList', () => {
  it('creates and removes, showing validation messages in Portuguese', async () => {
    renderWithApi(<SampleItemList />, testApi());
    expect(await screen.findByText('Nenhum registro ainda.')).toBeTruthy();

    create('   ');
    expect(await screen.findByText('Informe o nome.')).toBeTruthy();

    create('Primeiro');
    expect(await screen.findByText('Primeiro')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Remover Primeiro' }));
    await waitFor(() => expect(screen.queryByText('Primeiro')).toBeNull());
  });
});
