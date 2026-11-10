import { cleanup, fireEvent, screen, within } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { OrderDetails } from '../src/components/orders/order-details';
import { renderWithApi, testApi } from './api';

afterEach(cleanup);

function fillItem(productName: string, quantity: string, unitPrice: string) {
  const form = screen.getByRole('form', { name: 'Adicionar item' });
  fireEvent.change(within(form).getByLabelText('Produto'), { target: { value: productName } });
  fireEvent.change(within(form).getByLabelText('Quantidade'), { target: { value: quantity } });
  fireEvent.change(within(form).getByLabelText('Preço unitário'), { target: { value: unitPrice } });
  fireEvent.click(within(form).getByRole('button', { name: 'Adicionar item' }));
}

describe('OrderDetails', () => {
  it('adds items, places the order and shows the rule violations in Portuguese', async () => {
    const api = testApi();
    const order = await api.orders.create.mutate({ customerName: 'Ana' });
    renderWithApi(<OrderDetails orderId={order.id} />, api);

    expect(await screen.findByRole('heading', { name: 'Pedido de Ana' })).toBeTruthy();
    expect(screen.getByText('Rascunho')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Realizar pedido' }));
    expect(await screen.findByText('Um pedido sem itens não pode ser realizado.')).toBeTruthy();

    fillItem('Caneta', '0', '3,50');
    expect(await screen.findByText('A quantidade deve ser maior que zero.')).toBeTruthy();

    fillItem('Caneta', '2', '3,50');
    expect(await screen.findByText('Caneta')).toBeTruthy();
    expect(screen.getByTestId('order-total').textContent).toMatch(/7,00/);

    fireEvent.click(screen.getByRole('button', { name: 'Realizar pedido' }));
    expect(await screen.findByText('Realizado')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Realizar pedido' }));
    expect(await screen.findByText('O pedido já foi realizado.')).toBeTruthy();
  });

  it('says when the order does not exist', async () => {
    renderWithApi(<OrderDetails orderId={crypto.randomUUID()} />, testApi());
    expect(await screen.findByText('Pedido não encontrado.')).toBeTruthy();
  });
});
