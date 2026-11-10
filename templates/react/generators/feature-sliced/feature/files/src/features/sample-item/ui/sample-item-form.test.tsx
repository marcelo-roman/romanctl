import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { renderWithProviders, server } from '@tests/support';
import { SampleItemForm } from './sample-item-form';

describe('SampleItemForm', () => {
  it('sends the form and confirms it', async () => {
    let received: unknown;
    server.use(
      http.post('*/api/sample-item', async ({ request }) => {
        received = await request.json();
        return new HttpResponse(null, { status: 204 });
      }),
    );
    const { user } = renderWithProviders(<SampleItemForm />);

    await user.type(screen.getByLabelText('Nome'), ' Ana ');
    await user.click(screen.getByRole('button', { name: 'Enviar' }));

    expect(await screen.findByText('Enviado.')).toBeInTheDocument();
    expect(received).toEqual({ name: 'Ana' });
  });

  it('shows the detail of the problem the API returns', async () => {
    server.use(
      http.post('*/api/sample-item', () =>
        HttpResponse.json({ status: 422, title: 'Regra de negócio violada', detail: 'Não é possível agora.' }, { status: 422 }),
      ),
    );
    const { user } = renderWithProviders(<SampleItemForm />);

    await user.type(screen.getByLabelText('Nome'), 'Ana');
    await user.click(screen.getByRole('button', { name: 'Enviar' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('Não é possível agora.');
  });

  it('asks for the name before calling the API', async () => {
    const { user } = renderWithProviders(<SampleItemForm />);

    await user.click(screen.getByRole('button', { name: 'Enviar' }));

    expect(screen.getByRole('alert')).toHaveTextContent('Informe o nome.');
  });
});
