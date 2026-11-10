import { useCallback, useState } from 'react';
import { errorMessage } from '../bff/bff-client';

export interface Action {
  run: <T>(work: () => Promise<T>) => Promise<T | undefined>;
  error?: string;
  pending: boolean;
}

export function useAction(): Action {
  const [error, setError] = useState<string>();
  const [pending, setPending] = useState(false);

  const run = useCallback(async <T,>(work: () => Promise<T>): Promise<T | undefined> => {
    setPending(true);
    setError(undefined);
    try {
      return await work();
    } catch (failure) {
      setError(errorMessage(failure));
      return undefined;
    } finally {
      setPending(false);
    }
  }, []);

  return { run, error, pending };
}
