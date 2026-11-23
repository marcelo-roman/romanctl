import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { OrdersPage } from '../src/pages/OrdersPage';
import { mockFetch } from './fetch-mock';

const list = {
  orders: [{ id: 'o1', customerName: 'Ana', status: 'Draft', itemCount: 3, total: '$19.00' }],
};

describe('OrdersPage', () => {
  it('lists the orders the BFF prepared', async () => {
    mockFetch(() => ({ body: list }));

    render(<OrdersPage />);

    expect(await screen.findByRole('link', { name: 'Ana' })).toBeTruthy();
    expect(screen.getByText('Draft')).toBeTruthy();
    expect(screen.getByText('$19.00')).toBeTruthy();
  });

  it('creates an order and opens it', async () => {
    const calls = mockFetch((call) => (call.method === 'POST' ? { status: 201, body: { id: 'o2' } } : { body: { orders: [] } }));
    render(<OrdersPage />);
    await screen.findByText('No orders yet.');

    await userEvent.type(screen.getByLabelText('Customer name'), 'Bia');
    await userEvent.click(screen.getByRole('button', { name: 'Create order' }));

    expect(calls.at(-1)).toEqual({ method: 'POST', url: '/actions/orders', body: { customerName: 'Bia' } });
    expect(window.location.pathname).toBe('/orders/o2');
  });

  it('shows the message the BFF sends when something fails', async () => {
    mockFetch((call) =>
      call.method === 'POST' ? { status: 422, body: { message: 'The customer name is required.' } } : { body: { orders: [] } },
    );
    render(<OrdersPage />);
    await screen.findByText('No orders yet.');

    await userEvent.click(screen.getByRole('button', { name: 'Create order' }));

    expect((await screen.findByRole('alert')).textContent).toBe('The customer name is required.');
    expect(window.location.pathname).toBe('/');
  });
});
