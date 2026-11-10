import { describe, expect, it } from 'vitest';
import { money } from '@/entities/order';
import { emptyItemForm, previewSubtotal, toAddItemInput } from './item-form';

describe('item form', () => {
  it('turns the typed text into the API payload', () => {
    const input = toAddItemInput({ productName: ' Caneta ', quantity: '2', unitPrice: '3,50', currency: 'usd' });

    expect(input).toEqual({ productName: 'Caneta', quantity: 2, unitPrice: 3.5, currency: 'USD' });
  });

  it('sends zero for numbers it cannot read and leaves the currency to the API when blank', () => {
    const input = toAddItemInput({ ...emptyItemForm, quantity: 'x', unitPrice: '', currency: ' ' });

    expect(input).toEqual({ productName: '', quantity: 0, unitPrice: 0 });
  });

  it('previews the subtotal only for a positive quantity and a price', () => {
    expect(previewSubtotal({ ...emptyItemForm, quantity: '3', unitPrice: '0,10' })).toEqual(money(0.3));
    expect(previewSubtotal({ ...emptyItemForm, quantity: '0', unitPrice: '1' })).toBeUndefined();
    expect(previewSubtotal({ ...emptyItemForm, unitPrice: '' })).toBeUndefined();
  });
});
