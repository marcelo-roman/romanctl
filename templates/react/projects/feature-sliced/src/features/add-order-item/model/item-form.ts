import { defaultCurrency, money, multiplyMoney, type Money } from '@/entities/order';
import { parseDecimal, parseInteger } from '@/shared/lib';
import type { AddItemInput } from '../api/add-order-item';

export interface ItemFormValues {
  productName: string;
  quantity: string;
  unitPrice: string;
  currency: string;
}

export const emptyItemForm: ItemFormValues = {
  productName: '',
  quantity: '1',
  unitPrice: '',
  currency: defaultCurrency,
};

function finiteOrZero(value: number): number {
  return Number.isFinite(value) ? value : 0;
}

export function toAddItemInput(values: ItemFormValues): AddItemInput {
  const currency = values.currency.trim().toUpperCase();
  return {
    productName: values.productName.trim(),
    quantity: finiteOrZero(parseInteger(values.quantity)),
    unitPrice: finiteOrZero(parseDecimal(values.unitPrice)),
    ...(currency ? { currency } : {}),
  };
}

export function previewSubtotal(values: ItemFormValues): Money | undefined {
  const quantity = parseInteger(values.quantity);
  const unitPrice = parseDecimal(values.unitPrice);
  if (!Number.isFinite(quantity) || !Number.isFinite(unitPrice) || quantity <= 0) {
    return undefined;
  }
  return multiplyMoney(money(unitPrice, values.currency.trim() || defaultCurrency), quantity);
}
