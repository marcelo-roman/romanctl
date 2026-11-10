import { Link } from '../router';

export function NotFoundPage() {
  return (
    <section>
      <h1>Página não encontrada</h1>
      <Link to="/orders">Ir para os pedidos</Link>
    </section>
  );
}
