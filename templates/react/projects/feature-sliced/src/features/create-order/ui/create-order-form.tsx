import { Alert, Button, Field, Input } from '@/shared/ui';
import { useCreateOrderForm } from '../model/use-create-order-form';

export function CreateOrderForm() {
  const form = useCreateOrderForm();

  return (
    <form className="form form-inline" onSubmit={form.submit} aria-label="New order">
      <Field label="Customer" htmlFor="customer-name">
        <Input
          id="customer-name"
          value={form.customerName}
          onChange={(event) => form.setCustomerName(event.target.value)}
          placeholder="Customer name"
        />
      </Field>
      <Button type="submit" disabled={form.isPending}>
        {form.isPending ? 'Creating...' : 'Create order'}
      </Button>
      {form.error && <Alert tone="error">{form.error}</Alert>}
    </form>
  );
}
