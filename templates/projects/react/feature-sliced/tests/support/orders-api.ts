import { http, HttpResponse } from 'msw';

interface FakeItem {
  productName: string;
  quantity: number;
  unitPrice: number;
  subtotal: number;
}

interface FakeOrder {
  id: string;
  customerName: string;
  status: 'Draft' | 'Placed';
  createdAt: string;
  total: number;
  currency: string;
  items: FakeItem[];
}

interface AddItemBody {
  productName?: string;
  quantity?: number;
  unitPrice?: number;
  currency?: string;
}

const base = '*/api/orders';
const orders = new Map<string, FakeOrder>();
let sequence = 0;

function problem(status: number, title: string, detail: string) {
  return HttpResponse.json({ status, title, detail }, { status });
}

const notFound = () => problem(404, 'Não encontrado', 'não encontrado');
const rule = (detail: string) => problem(422, 'Regra de negócio violada', detail);

function cents(value: number): number {
  return Math.round(value * 100);
}

function addItem(order: FakeOrder, body: AddItemBody): FakeOrder {
  const quantity = body.quantity ?? 0;
  const unitPrice = body.unitPrice ?? 0;
  const item = { productName: body.productName ?? '', quantity, unitPrice, subtotal: (cents(unitPrice) * quantity) / 100 };
  const items = [...order.items, item];
  return { ...order, items, total: items.reduce((sum, it) => sum + cents(it.subtotal), 0) / 100 };
}

function itemViolation(order: FakeOrder, body: AddItemBody): string | undefined {
  if (order.status === 'Placed') {
    return 'Um pedido já realizado não pode ser alterado.';
  }
  if (!body.productName?.trim()) {
    return 'O nome do produto é obrigatório.';
  }
  if (!body.quantity || body.quantity <= 0) {
    return 'A quantidade deve ser maior que zero.';
  }
  return undefined;
}

export function seedOrder(customerName: string, changes: Partial<FakeOrder> = {}): FakeOrder {
  sequence += 1;
  const order: FakeOrder = {
    id: `order-${sequence}`,
    customerName,
    status: 'Draft',
    createdAt: new Date(Date.UTC(2026, 0, sequence)).toISOString(),
    total: 0,
    currency: 'BRL',
    items: [],
    ...changes,
  };
  orders.set(order.id, order);
  return order;
}

export function resetOrders(): void {
  orders.clear();
  sequence = 0;
}

export const ordersApi = [
  http.post(base, async ({ request }) => {
    const body = (await request.json()) as { customerName?: string };
    if (!body.customerName?.trim()) {
      return rule('O nome do cliente é obrigatório.');
    }
    return HttpResponse.json({ id: seedOrder(body.customerName.trim()).id }, { status: 201 });
  }),
  http.post(`${base}/:id/items`, async ({ params, request }) => {
    const order = orders.get(String(params.id));
    if (!order) {
      return notFound();
    }
    const body = (await request.json()) as AddItemBody;
    const violation = itemViolation(order, body);
    if (violation) {
      return rule(violation);
    }
    orders.set(order.id, addItem(order, body));
    return new HttpResponse(null, { status: 204 });
  }),
  http.post(`${base}/:id/place`, ({ params }) => {
    const order = orders.get(String(params.id));
    if (!order) {
      return notFound();
    }
    if (order.status === 'Placed') {
      return rule('O pedido já foi realizado.');
    }
    if (order.items.length === 0) {
      return rule('Um pedido sem itens não pode ser realizado.');
    }
    orders.set(order.id, { ...order, status: 'Placed' });
    return new HttpResponse(null, { status: 204 });
  }),
  http.get(`${base}/:id`, ({ params }) => {
    const order = orders.get(String(params.id));
    return order ? HttpResponse.json(order) : notFound();
  }),
  http.get(base, ({ request }) => {
    const limit = Number(new URL(request.url).searchParams.get('limit') ?? 50);
    return HttpResponse.json([...orders.values()].reverse().slice(0, limit));
  }),
];
