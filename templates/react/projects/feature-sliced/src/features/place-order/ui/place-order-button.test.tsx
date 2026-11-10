import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { statusLabel, useOrder } from '@/entities/order';
import { renderWithProviders, seedOrder } from '@tests/support';
import { PlaceOrderButton } from './place-order-button';

function OrderStatus({ id }: { id: string }) {
  const { data } = useOrder(id);
  return data ? <output>{`Status: ${statusLabel(data.status)}`}</output> : null;
}

function renderButton(orderId: string) {
  return renderWithProviders(
    <>
      <PlaceOrderButton orderId={orderId} />
      <OrderStatus id={orderId} />
    </>,
  );
}

describe('PlaceOrderButton', () => {
  it('places the order and the status follows', async () => {
    const order = seedOrder('Ana', { items: [{ productName: 'Caneta', quantity: 1, unitPrice: 2, subtotal: 2 }], total: 2 });
    const { user } = renderButton(order.id);
    expect(await screen.findByText('Status: Rascunho')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Realizar pedido' }));

    expect(await screen.findByText('Status: Realizado')).toBeInTheDocument();
  });

  it('explains why an empty order cannot be placed', async () => {
    const order = seedOrder('Ana');
    const { user } = renderButton(order.id);

    await user.click(screen.getByRole('button', { name: 'Realizar pedido' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('Um pedido sem itens não pode ser realizado.');
  });
});
