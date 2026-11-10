import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { OrderPage } from '../src/pages/OrderPage';
import { mockFetch } from './fetch-mock';

function detail(overrides: Record<string, unknown> = {}, items: unknown[] = []) {
  return {
    order: { id: 'o1', customerName: 'Ana', status: 'Rascunho', total: 'R$ 0,00', canAddItems: true, canPlace: false, items, ...overrides },
    suggestions: [
      { id: 'p1', name: 'Caderno', price: 'R$ 12,00' },
      { id: 'p2', name: 'Caneta', price: 'R$ 3,50' },
    ],
  };
}

const withItem = detail({ total: 'R$ 7,00', canPlace: true }, [
  { id: 'i1', productName: 'Caneta', quantity: 2, unitPrice: 'R$ 3,50', subtotal: 'R$ 7,00' },
]);

describe('OrderPage', () => {
  it('adds the chosen product and shows the order the BFF answers', async () => {
    const calls = mockFetch((call) => ({ body: call.method === 'POST' ? withItem : detail() }));
    render(<OrderPage id="o1" />);
    await screen.findByText('Pedido de Ana');

    await userEvent.selectOptions(screen.getByLabelText('Produto'), 'p2');
    await userEvent.clear(screen.getByLabelText('Quantidade'));
    await userEvent.type(screen.getByLabelText('Quantidade'), '2');
    await userEvent.click(screen.getByRole('button', { name: 'Adicionar' }));

    expect(calls.at(-1)).toEqual({ method: 'POST', url: '/actions/orders/o1/items', body: { productId: 'p2', quantity: 2 } });
    expect(await screen.findByText('R$ 7,00', { selector: 'tfoot td' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Realizar pedido' })).toBeTruthy();
  });

  it('places the order and stops offering changes', async () => {
    const placed = detail({ status: 'Realizado', canAddItems: false, canPlace: false, total: 'R$ 7,00' }, withItem.order.items);
    mockFetch((call) => ({ body: call.method === 'POST' ? { ...placed, suggestions: [] } : withItem }));
    render(<OrderPage id="o1" />);

    await userEvent.click(await screen.findByRole('button', { name: 'Realizar pedido' }));

    expect(await screen.findByText('Situação: Realizado')).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Adicionar' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Realizar pedido' })).toBeNull();
  });

  it('shows the rule the API refused', async () => {
    mockFetch((call) =>
      call.method === 'POST' ? { status: 422, body: { message: 'A quantidade deve ser maior que zero.' } } : { body: detail() },
    );
    render(<OrderPage id="o1" />);
    await screen.findByText('Pedido de Ana');

    await userEvent.clear(screen.getByLabelText('Quantidade'));
    await userEvent.type(screen.getByLabelText('Quantidade'), '0');
    await userEvent.click(screen.getByRole('button', { name: 'Adicionar' }));

    expect((await screen.findByRole('alert')).textContent).toBe('A quantidade deve ser maior que zero.');
  });

  it('explains when the order does not exist', async () => {
    mockFetch(() => ({ status: 404, body: { message: 'Pedido não encontrado.' } }));
    render(<OrderPage id="nada" />);

    expect((await screen.findByRole('alert')).textContent).toBe('Pedido não encontrado.');
  });
});
