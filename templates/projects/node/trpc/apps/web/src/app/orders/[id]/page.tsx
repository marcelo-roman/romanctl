import Link from 'next/link';
import { OrderDetails } from '../../../components/orders/order-details';

interface OrderPageProps {
  params: Promise<{ id: string }>;
}

export default async function OrderPage({ params }: OrderPageProps) {
  const { id } = await params;
  return (
    <>
      <Link href="/">Voltar para os pedidos</Link>
      <OrderDetails orderId={id} />
    </>
  );
}
