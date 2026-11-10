import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '@tests/support';
import { SampleItemPage } from './sample-item-page';

describe('SampleItemPage', () => {
  it('shows its title', () => {
    renderWithProviders(<SampleItemPage />);

    expect(screen.getByRole('heading', { name: 'Sample Item' })).toBeInTheDocument();
  });
});
