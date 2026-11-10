'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { FormEvent } from 'react';
import { errorMessage } from '../../lib/error-message';
import { formatDate } from '../../lib/format';
import { useTRPC } from '../../trpc/react';

export function SampleItemList() {
  const trpc = useTRPC();
  const queryClient = useQueryClient();
  const sampleItems = useQuery(trpc.sampleItems.list.queryOptions());
  const refresh = () => queryClient.invalidateQueries({ queryKey: trpc.sampleItems.list.queryKey() });
  const create = useMutation(trpc.sampleItems.create.mutationOptions({ onSuccess: refresh }));
  const remove = useMutation(trpc.sampleItems.remove.mutationOptions({ onSuccess: refresh }));

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formElement = event.currentTarget;
    const form = new FormData(formElement);
    create.mutate({ name: String(form.get('name') ?? '') }, { onSuccess: () => formElement.reset() });
  }

  return (
    <section>
      <form onSubmit={submit} aria-label="Novo registro">
        <label>
          Nome
          <input name="name" autoComplete="off" />
        </label>
        <button type="submit" disabled={create.isPending}>
          Criar
        </button>
        {create.isError && <p role="alert">{errorMessage(create.error)}</p>}
      </form>
      {remove.isError && <p role="alert">{errorMessage(remove.error)}</p>}
      {sampleItems.isPending && <p>Carregando...</p>}
      {sampleItems.isError && <p role="alert">{errorMessage(sampleItems.error)}</p>}
      {sampleItems.data?.length === 0 && <p>Nenhum registro ainda.</p>}
      {sampleItems.data && sampleItems.data.length > 0 && (
        <table>
          <thead>
            <tr>
              <th>Nome</th>
              <th>Criado em</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {sampleItems.data.map((sampleItem) => (
              <tr key={sampleItem.id}>
                <td>{sampleItem.name}</td>
                <td>{formatDate(sampleItem.createdAt)}</td>
                <td>
                  <button type="button" disabled={remove.isPending} onClick={() => remove.mutate({ id: sampleItem.id })}>
                    Remover {sampleItem.name}
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
