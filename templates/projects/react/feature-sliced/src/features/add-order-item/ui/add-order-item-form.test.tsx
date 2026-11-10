import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { formatMoney, useOrder } from '@/entities/order';
import { renderWithProviders, seedOrder } from '@tests/support';
import { AddOrderItemForm } from './add-order-item-form';

function OrderTotal({ id }: { id: string }) {
  const { data } = useOrder(id);
  return data ? <output>{`Total ${formatMoney(data.total)}`}</output> : null;
}

function renderForm() {
  const order = seedOrder('Ana');
  const view = renderWithProviders(
    <>
      <AddOrderItemForm orderId={order.id} />
      <OrderTotal id={order.id} />
    </>,
  );
  return { ...view, order };
}

describe('AddOrderItemForm', () => {
  it('adds the item and refreshes the order total', async () => {
    const { user } = renderForm();

    await user.type(screen.getByLabelText('Produto'), 'Caneta');
    await user.clear(screen.getByLabelText('Quantidade'));
    await user.type(screen.getByLabelText('Quantidade'), '2');
    await user.type(screen.getByLabelText('Preço unitário'), '3,50');
    expect(screen.getByText(/Subtotal: R\$\s7,00/)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Adicionar item' }));

    expect(await screen.findByText(/Total R\$\s7,00/)).toBeInTheDocument();
    expect(screen.getByLabelText('Produto')).toHaveValue('');
  });

  it('shows the 422 detail when the quantity is zero', async () => {
    const { user } = renderForm();

    await user.type(screen.getByLabelText('Produto'), 'Caneta');
    await user.clear(screen.getByLabelText('Quantidade'));
    await user.type(screen.getByLabelText('Quantidade'), '0');
    await user.type(screen.getByLabelText('Preço unitário'), '3,50');
    await user.click(screen.getByRole('button', { name: 'Adicionar item' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('A quantidade deve ser maior que zero.');
  });
});
