import { Link } from '../router';

export function NotFoundPage() {
  return (
    <section>
      <h1>Page not found</h1>
      <Link to="/orders">Go to the orders</Link>
    </section>
  );
}
