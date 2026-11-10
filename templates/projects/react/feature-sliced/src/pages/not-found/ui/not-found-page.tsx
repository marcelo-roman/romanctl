import { Link } from 'react-router';
import { paths } from '@/shared/config';

export function NotFoundPage() {
  return (
    <main className="page">
      <h1>Página não encontrada</h1>
      <p className="muted">O endereço não corresponde a nenhuma página.</p>
      <Link to={paths.orders}>Voltar para os pedidos</Link>
    </main>
  );
}
