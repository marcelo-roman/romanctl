import { Alert, Button, Field, Input } from '@/shared/ui';
import { useAddOrderItemForm } from '../model/use-add-order-item-form';

export function AddOrderItemForm({ orderId }: { orderId: string }) {
  const form = useAddOrderItemForm(orderId);

  return (
    <form className="form" onSubmit={form.submit} aria-label="Add item">
      <div className="form-grid">
        <Field label="Product" htmlFor="item-product">
          <Input id="item-product" value={form.values.productName} onChange={(e) => form.change('productName')(e.target.value)} />
        </Field>
        <Field label="Quantity" htmlFor="item-quantity">
          <Input id="item-quantity" inputMode="numeric" value={form.values.quantity} onChange={(e) => form.change('quantity')(e.target.value)} />
        </Field>
        <Field label="Unit price" htmlFor="item-price" hint={form.subtotal && `Subtotal: ${form.subtotal}`}>
          <Input id="item-price" inputMode="decimal" value={form.values.unitPrice} onChange={(e) => form.change('unitPrice')(e.target.value)} />
        </Field>
        <Field label="Currency" htmlFor="item-currency">
          <Input id="item-currency" maxLength={3} value={form.values.currency} onChange={(e) => form.change('currency')(e.target.value)} />
        </Field>
      </div>
      <Button type="submit" disabled={form.isPending}>
        {form.isPending ? 'Adding...' : 'Add item'}
      </Button>
      {form.error && <Alert tone="error">{form.error}</Alert>}
    </form>
  );
}
