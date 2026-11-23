import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { OrderPage } from '../src/pages/OrderPage';
import { mockFetch } from './fetch-mock';

function detail(overrides: Record<string, unknown> = {}, items: unknown[] = []) {
  return {
    order: { id: 'o1', customerName: 'Ana', status: 'Draft', total: '$0.00', canAddItems: true, canPlace: false, items, ...overrides },
    suggestions: [
      { id: 'p1', name: 'Notebook', price: '$12.00' },
      { id: 'p2', name: 'Pen', price: '$3.50' },
    ],
  };
}

const withItem = detail({ total: '$7.00', canPlace: true }, [
  { id: 'i1', productName: 'Pen', quantity: 2, unitPrice: '$3.50', subtotal: '$7.00' },
]);

describe('OrderPage', () => {
  it('adds the chosen product and shows the order the BFF answers', async () => {
    const calls = mockFetch((call) => ({ body: call.method === 'POST' ? withItem : detail() }));
    render(<OrderPage id="o1" />);
    await screen.findByText('Order for Ana');

    await userEvent.selectOptions(screen.getByLabelText('Product'), 'p2');
    await userEvent.clear(screen.getByLabelText('Quantity'));
    await userEvent.type(screen.getByLabelText('Quantity'), '2');
    await userEvent.click(screen.getByRole('button', { name: 'Add' }));

    expect(calls.at(-1)).toEqual({ method: 'POST', url: '/actions/orders/o1/items', body: { productId: 'p2', quantity: 2 } });
    expect(await screen.findByText('$7.00', { selector: 'tfoot td' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Place order' })).toBeTruthy();
  });

  it('places the order and stops offering changes', async () => {
    const placed = detail({ status: 'Placed', canAddItems: false, canPlace: false, total: '$7.00' }, withItem.order.items);
    mockFetch((call) => ({ body: call.method === 'POST' ? { ...placed, suggestions: [] } : withItem }));
    render(<OrderPage id="o1" />);

    await userEvent.click(await screen.findByRole('button', { name: 'Place order' }));

    expect(await screen.findByText('Status: Placed')).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Add' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Place order' })).toBeNull();
  });

  it('shows the rule the API refused', async () => {
    mockFetch((call) =>
      call.method === 'POST' ? { status: 422, body: { message: 'The quantity must be greater than zero.' } } : { body: detail() },
    );
    render(<OrderPage id="o1" />);
    await screen.findByText('Order for Ana');

    await userEvent.clear(screen.getByLabelText('Quantity'));
    await userEvent.type(screen.getByLabelText('Quantity'), '0');
    await userEvent.click(screen.getByRole('button', { name: 'Add' }));

    expect((await screen.findByRole('alert')).textContent).toBe('The quantity must be greater than zero.');
  });

  it('explains when the order does not exist', async () => {
    mockFetch(() => ({ status: 404, body: { message: 'Order not found.' } }));
    render(<OrderPage id="nothing" />);

    expect((await screen.findByRole('alert')).textContent).toBe('Order not found.');
  });
});
