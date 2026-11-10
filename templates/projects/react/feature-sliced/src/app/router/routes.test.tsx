import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderRoutes, seedOrder } from '@tests/support';
import { routes } from './routes';

describe('routes', () => {
  it('lists the recent orders on the home page and opens one', async () => {
    seedOrder('Ana', { items: [{ productName: 'Caneta', quantity: 2, unitPrice: 3.5, subtotal: 7 }], total: 7 });
    const { user } = renderRoutes(routes, '/');

    expect(screen.getByRole('heading', { name: 'Pedidos' })).toBeInTheDocument();
    await user.click(await screen.findByRole('link', { name: 'Ana' }));

    const summary = await screen.findByRole('region', { name: 'Resumo do pedido' });
    expect(within(summary).getByText('Caneta')).toBeInTheDocument();
    expect(within(summary).getByText(/R\$\s7,00/, { selector: 'strong' })).toBeInTheDocument();
  });

  it('hides the draft actions once the order is placed', async () => {
    const order = seedOrder('Bia', { status: 'Placed' });
    renderRoutes(routes, `/orders/${order.id}`);

    expect(await screen.findByText('Realizado')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Realizar pedido' })).not.toBeInTheDocument();
  });

  it('shows the API detail when the order does not exist', async () => {
    renderRoutes(routes, '/orders/missing');

    expect(await screen.findByRole('alert')).toHaveTextContent('não encontrado');
  });

  it('answers unknown addresses with the not found page', () => {
    renderRoutes(routes, '/nada-aqui');

    expect(screen.getByRole('heading', { name: 'Página não encontrada' })).toBeInTheDocument();
  });
});
