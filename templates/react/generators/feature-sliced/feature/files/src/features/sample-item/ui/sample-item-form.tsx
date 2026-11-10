import { Alert, Button, Field, Input } from '@/shared/ui';
import { useSampleItemForm } from '../model/use-sample-item-form';

export function SampleItemForm() {
  const form = useSampleItemForm();

  return (
    <form className="form" onSubmit={form.submit} aria-label="Sample Item">
      <Field label="Nome" htmlFor="sample-item-name">
        <Input id="sample-item-name" value={form.values.name} onChange={(event) => form.change('name')(event.target.value)} />
      </Field>
      <Button type="submit" disabled={form.isPending}>
        {form.isPending ? 'Enviando...' : 'Enviar'}
      </Button>
      {form.error && <Alert tone="error">{form.error}</Alert>}
      {form.succeeded && <Alert tone="success">Enviado.</Alert>}
    </form>
  );
}
