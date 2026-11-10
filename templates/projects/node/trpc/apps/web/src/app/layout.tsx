import type { Metadata } from 'next';
import Link from 'next/link';
import type { ReactNode } from 'react';
import { ApiProvider } from '../trpc/api-provider';
import './globals.css';

export const metadata: Metadata = {
  title: 'Sample App',
  description: 'Pedidos com Next.js e tRPC',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="pt-BR">
      <body>
        <header>
          <nav>
            <Link href="/">Pedidos</Link>
            {/* romanctl:links */}
          </nav>
        </header>
        <main>
          <ApiProvider>{children}</ApiProvider>
        </main>
      </body>
    </html>
  );
}
