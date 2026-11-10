import { memoryRepositories } from '@sample-app/db';
import { TRPCError } from '@trpc/server';
import { describe, expect, it } from 'vitest';
import { createContext } from '../../src/context';
import { DomainError } from '../../src/domain/domain-error';
import { SampleItem } from '../../src/domain/sample-item';
import { createCaller } from '../../src/trpc/router';

function caller() {
  return createCaller(createContext(memoryRepositories()));
}

async function code(promise: Promise<unknown>): Promise<string | undefined> {
  const error = await promise.then(
    () => undefined,
    (reason: unknown) => reason,
  );
  return error instanceof TRPCError ? error.code : undefined;
}

describe('SampleItem', () => {
  it('requires a name of at most 200 characters', () => {
    expect(() => SampleItem.create('  ', new Date())).toThrow(DomainError);
    expect(() => SampleItem.create('x'.repeat(201), new Date())).toThrow(DomainError);
    expect(SampleItem.create(' Primeiro ', new Date()).name).toBe('Primeiro');
  });
});

describe('sample items router', () => {
  it('creates, renames, lists and removes', async () => {
    const api = caller();
    const created = await api.sampleItems.create({ name: 'Primeiro' });
    const renamed = await api.sampleItems.rename({ id: created.id, name: 'Renomeado' });

    expect(renamed.name).toBe('Renomeado');
    expect(await api.sampleItems.byId({ id: created.id })).toEqual(renamed);
    expect((await api.sampleItems.list()).map((sampleItem) => sampleItem.id)).toEqual([created.id]);

    await api.sampleItems.remove({ id: created.id });
    expect(await code(api.sampleItems.byId({ id: created.id }))).toBe('NOT_FOUND');
    expect(await code(api.sampleItems.remove({ id: created.id }))).toBe('NOT_FOUND');
  });

  it('answers BAD_REQUEST for invalid input', async () => {
    const api = caller();
    expect(await code(api.sampleItems.create({ name: '  ' }))).toBe('BAD_REQUEST');
    expect(await code(api.sampleItems.byId({ id: '42' }))).toBe('BAD_REQUEST');
  });
});
