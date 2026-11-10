import { Alert, Button, Field, Input } from '@/shared/ui';
import { useCreateOrderForm } from '../model/use-create-order-form';

export function CreateOrderForm() {
  const form = useCreateOrderForm();

  return (
    <form className="form form-inline" onSubmit={form.submit} aria-label="Novo pedido">
      <Field label="Cliente" htmlFor="customer-name">
        <Input
          id="customer-name"
          value={form.customerName}
          onChange={(event) => form.setCustomerName(event.target.value)}
          placeholder="Nome do cliente"
        />
      </Field>
      <Button type="submit" disabled={form.isPending}>
        {form.isPending ? 'Criando...' : 'Criar pedido'}
      </Button>
      {form.error && <Alert tone="error">{form.error}</Alert>}
    </form>
  );
}
