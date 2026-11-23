import { screen } from '@testing-library/react';
import { useParams } from 'react-router';
import { describe, expect, it } from 'vitest';
import { statusLabel, useOrder } from '@/entities/order';
import { renderRoutes } from '@tests/support';
import { CreateOrderForm } from './create-order-form';

function OrderProbe() {
  const { id = '' } = useParams();
  const { data } = useOrder(id);
  return data ? <p>{`${data.customerName}: ${statusLabel(data.status)}`}</p> : null;
}

const routes = [
  { path: '/', element: <CreateOrderForm /> },
  { path: '/orders/:id', element: <OrderProbe /> },
];

describe('CreateOrderForm', () => {
  it('creates the order and shows it', async () => {
    const { user, router } = renderRoutes(routes);

    await user.type(screen.getByLabelText('Customer'), 'Ana');
    await user.click(screen.getByRole('button', { name: 'Create order' }));

    expect(await screen.findByText('Ana: Draft')).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/orders/order-1');
  });

  it('shows the detail of the problem the API returns', async () => {
    const { user } = renderRoutes(routes);

    await user.click(screen.getByRole('button', { name: 'Create order' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('The customer name is required.');
  });
});
