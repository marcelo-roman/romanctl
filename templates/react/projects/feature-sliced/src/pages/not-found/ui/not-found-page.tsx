import { Link } from 'react-router';
import { paths } from '@/shared/config';

export function NotFoundPage() {
  return (
    <main className="page">
      <h1>Page not found</h1>
      <p className="muted">The address does not match any page.</p>
      <Link to={paths.orders}>Back to orders</Link>
    </main>
  );
}
