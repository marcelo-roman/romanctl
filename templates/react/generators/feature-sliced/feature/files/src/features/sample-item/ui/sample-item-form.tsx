import { Alert, Button, Field, Input } from '@/shared/ui';
import { useSampleItemForm } from '../model/use-sample-item-form';

export function SampleItemForm() {
  const form = useSampleItemForm();

  return (
    <form className="form" onSubmit={form.submit} aria-label="Sample Item">
      <Field label="Name" htmlFor="sample-item-name">
        <Input id="sample-item-name" value={form.values.name} onChange={(event) => form.change('name')(event.target.value)} />
      </Field>
      <Button type="submit" disabled={form.isPending}>
        {form.isPending ? 'Submitting...' : 'Submit'}
      </Button>
      {form.error && <Alert tone="error">{form.error}</Alert>}
      {form.succeeded && <Alert tone="success">Submitted.</Alert>}
    </form>
  );
}
