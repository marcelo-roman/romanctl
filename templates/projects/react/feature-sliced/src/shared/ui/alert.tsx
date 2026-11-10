import type { ReactNode } from 'react';

type AlertTone = 'error' | 'success' | 'info';

interface AlertProps {
  tone?: AlertTone;
  children: ReactNode;
}

export function Alert({ tone = 'info', children }: AlertProps) {
  return (
    <div role={tone === 'error' ? 'alert' : 'status'} className={`alert alert-${tone}`}>
      {children}
    </div>
  );
}
