import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { App } from '../src/App';
import { mockFetch } from './fetch-mock';

describe('App', () => {
  it('routes between the list and the order', async () => {
    mockFetch((call) =>
      call.url === '/screens/orders'
        ? { body: { orders: [{ id: 'o1', customerName: 'Ana', status: 'Rascunho', itemCount: 0, total: 'R$ 0,00' }] } }
        : {
            body: {
              order: { id: 'o1', customerName: 'Ana', status: 'Rascunho', total: 'R$ 0,00', canAddItems: false, canPlace: false, items: [] },
              suggestions: [],
            },
          },
    );
    render(<App />);

    await userEvent.click(await screen.findByRole('link', { name: 'Ana' }));

    expect(await screen.findByText('Pedido de Ana')).toBeTruthy();
    expect(window.location.pathname).toBe('/orders/o1');
  });

  it('answers unknown pages', () => {
    window.history.replaceState(null, '', '/nada');
    render(<App />);
    expect(screen.getByText('Página não encontrada')).toBeTruthy();
  });
});
