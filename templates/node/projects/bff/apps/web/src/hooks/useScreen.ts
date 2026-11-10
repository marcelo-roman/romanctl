import { useCallback, useEffect, useState } from 'react';
import { errorMessage } from '../bff/bff-client';

interface State<T> {
  data?: T;
  error?: string;
  loading: boolean;
}

export interface Screen<T> extends State<T> {
  replace: (data: T) => void;
  reload: () => void;
}

export function useScreen<T>(key: string, load: () => Promise<T>): Screen<T> {
  const [state, setState] = useState<State<T>>({ loading: true });
  const [version, setVersion] = useState(0);

  useEffect(() => {
    let active = true;
    setState((current) => ({ data: current.data, loading: true }));
    load().then(
      (data) => active && setState({ data, loading: false }),
      (error: unknown) => active && setState({ error: errorMessage(error), loading: false }),
    );
    return () => {
      active = false;
    };
  }, [key, version]);

  const replace = useCallback((data: T) => setState({ data, loading: false }), []);
  const reload = useCallback(() => setVersion((current) => current + 1), []);
  return { ...state, replace, reload };
}
