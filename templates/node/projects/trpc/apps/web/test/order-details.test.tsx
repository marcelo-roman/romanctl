import { cleanup, fireEvent, screen, within } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { OrderDetails } from '../src/components/orders/order-details';
import { renderWithApi, testApi } from './api';

afterEach(cleanup);

function fillItem(productName: string, quantity: string, unitPrice: string) {
  const form = screen.getByRole('form', { name: 'Add item' });
  fireEvent.change(within(form).getByLabelText('Product'), { target: { value: productName } });
  fireEvent.change(within(form).getByLabelText('Quantity'), { target: { value: quantity } });
  fireEvent.change(within(form).getByLabelText('Unit price'), { target: { value: unitPrice } });
  fireEvent.click(within(form).getByRole('button', { name: 'Add item' }));
}

describe('OrderDetails', () => {
  it('adds items, places the order and shows the rule violations in English', async () => {
    const api = testApi();
    const order = await api.orders.create.mutate({ customerName: 'Ana' });
    renderWithApi(<OrderDetails orderId={order.id} />, api);

    expect(await screen.findByRole('heading', { name: 'Order for Ana' })).toBeTruthy();
    expect(screen.getByText('Draft')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Place order' }));
    expect(await screen.findByText('An order without items cannot be placed.')).toBeTruthy();

    fillItem('Pen', '0', '3.50');
    expect(await screen.findByText('The quantity must be greater than zero.')).toBeTruthy();

    fillItem('Pen', '2', '3.50');
    expect(await screen.findByText('Pen')).toBeTruthy();
    expect(screen.getByTestId('order-total').textContent).toMatch(/7\.00/);

    fireEvent.click(screen.getByRole('button', { name: 'Place order' }));
    expect(await screen.findByText('Placed')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Place order' }));
    expect(await screen.findByText('The order has already been placed.')).toBeTruthy();
  });

  it('says when the order does not exist', async () => {
    renderWithApi(<OrderDetails orderId={crypto.randomUUID()} />, testApi());
    expect(await screen.findByText('Order not found.')).toBeTruthy();
  });
});
