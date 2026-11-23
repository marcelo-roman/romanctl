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

    await user.type(screen.getByLabelText('Name'), ' Ana ');
    await user.click(screen.getByRole('button', { name: 'Submit' }));

    expect(await screen.findByText('Submitted.')).toBeInTheDocument();
    expect(received).toEqual({ name: 'Ana' });
  });

  it('shows the detail of the problem the API returns', async () => {
    server.use(
      http.post('*/api/sample-item', () =>
        HttpResponse.json({ status: 422, title: 'Business rule violated', detail: 'Not possible right now.' }, { status: 422 }),
      ),
    );
    const { user } = renderWithProviders(<SampleItemForm />);

    await user.type(screen.getByLabelText('Name'), 'Ana');
    await user.click(screen.getByRole('button', { name: 'Submit' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('Not possible right now.');
  });

  it('asks for the name before calling the API', async () => {
    const { user } = renderWithProviders(<SampleItemForm />);

    await user.click(screen.getByRole('button', { name: 'Submit' }));

    expect(screen.getByRole('alert')).toHaveTextContent('Enter the name.');
  });
});
