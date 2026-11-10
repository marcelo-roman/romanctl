import { type FormEvent, useState } from 'react';

interface NewOrderFormProps {
  pending: boolean;
  onCreate: (customerName: string) => Promise<void>;
}

export function NewOrderForm({ pending, onCreate }: NewOrderFormProps) {
  const [customerName, setCustomerName] = useState('');

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onCreate(customerName);
  }

  return (
    <form className="inline-form" onSubmit={submit}>
      <label>
        Nome do cliente
        <input value={customerName} onChange={(event) => setCustomerName(event.target.value)} />
      </label>
      <button type="submit" disabled={pending}>
        Criar pedido
      </button>
    </form>
  );
}
