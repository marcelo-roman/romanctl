import { Layout } from './components/Layout';
import { usePath } from './router';
import { renderRoute } from './routes';

export function App() {
  const path = usePath();
  return <Layout>{renderRoute(path)}</Layout>;
}
