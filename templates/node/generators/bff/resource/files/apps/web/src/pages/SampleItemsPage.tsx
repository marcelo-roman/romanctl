import { type FormEvent, useState } from 'react';
import { type SampleItemsScreen, sampleItemsBff } from '../bff/sample-items';
import { ErrorMessage } from '../components/ErrorMessage';
import { useAction } from '../hooks/useAction';
import { useScreen } from '../hooks/useScreen';

export function SampleItemsPage() {
  const screen = useScreen('sample-items', sampleItemsBff.list);
  const action = useAction();
  const [name, setName] = useState('');

  async function change(work: () => Promise<SampleItemsScreen>) {
    const next = await action.run(work);
    if (next) {
      screen.replace(next);
    }
    return next;
  }

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (await change(() => sampleItemsBff.create(name))) {
      setName('');
    }
  }

  const sampleItems = screen.data?.sampleItems ?? [];
  return (
    <section>
      <h1>Sample Items</h1>
      <form className="inline-form" onSubmit={create}>
        <label>
          Name
          <input value={name} onChange={(event) => setName(event.target.value)} />
        </label>
        <button type="submit" disabled={action.pending}>
          Add
        </button>
      </form>
      <ErrorMessage message={action.error ?? screen.error} />
      {screen.loading && !screen.data && <p>Loading...</p>}
      {screen.data && sampleItems.length === 0 && <p>No items yet.</p>}
      {sampleItems.length > 0 && (
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Created on</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {sampleItems.map((sampleItem) => (
              <tr key={sampleItem.id}>
                <td>{sampleItem.name}</td>
                <td>{sampleItem.createdOn}</td>
                <td>
                  <button type="button" disabled={action.pending} onClick={() => change(() => sampleItemsBff.delete(sampleItem.id))}>
                    Remove {sampleItem.name}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
