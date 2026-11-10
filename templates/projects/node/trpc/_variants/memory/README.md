# SampleApp

Monorepo full-stack com Next.js e tRPC, em TypeScript, gerado pelo `romanctl`.

```bash
romanctl run sample-app
npm test
```

Endereço: http://localhost:3200. A porta muda com a variável `PORT`.

| Script | O que faz |
|---|---|
| `npm run dev` | Sobe o Next.js em modo de desenvolvimento, com a API tRPC em `/api/trpc` |
| `npm test` | Confere os tipos e roda os testes de todos os pacotes com Vitest |
| `npm run build` | Gera o build de produção do Next.js, com `output: "standalone"` |
| `npm start` | Sobe o build de produção |

## Pacotes

Um só processo: o Next.js serve as páginas e a API tRPC, por um route handler em
`apps/web/src/app/api/trpc/[trpc]/route.ts`. O teste em `tests/architecture.test.ts` garante a direção das
dependências.

| Pasta | Conteúdo | Depende de |
|---|---|---|
| `packages/db` | Interfaces dos repositórios, com registros simples, e a implementação em `memory/` | nada |
| `packages/api` | Domínio (`Order`, `Money`), serviços com as regras e o router tRPC com validação por zod | as interfaces de `db`, só com `import type` |
| `apps/web` | Next.js (App Router), React e TanStack Query com `@trpc/tanstack-react-query` | o tipo `AppRouter` e o cliente tRPC |

Em `apps/web`, só `src/server` e o route handler importam valores de `@sample-app/api` e escolhem os repositórios.
Os componentes conhecem a API apenas pelo tipo `AppRouter`, pelo cliente tRPC. O domínio de `packages/api` não
depende de nada fora dele, e `packages/api` não conhece Next.js nem React.

## Procedimentos

| Procedimento | Tipo | O que faz |
|---|---|---|
| `orders.create` | mutation | Cria um pedido: `{ customerName: "Ana" }` |
| `orders.addItem` | mutation | Adiciona um item: `{ orderId, productName: "Caneta", quantity: 2, unitPrice: 3.5, currency: "BRL" }` |
| `orders.place` | mutation | Realiza o pedido: `{ orderId }` |
| `orders.byId` | query | Um pedido: `{ orderId }` |
| `orders.list` | query | Os pedidos mais recentes; `limit` vale 50 se omitido e no máximo 200 |

```bash
curl 'http://localhost:3200/api/trpc/orders.list'
```

## Regras

- O nome do cliente e o nome do produto são obrigatórios.
- A quantidade é um número inteiro maior que zero; o preço não pode ser negativo; a moeda padrão é BRL.
- Todos os itens de um pedido usam a mesma moeda.
- Só um pedido em rascunho, com pelo menos um item, pode ser realizado; depois disso, ele não muda mais.

| Código tRPC | HTTP | Quando |
|---|---|---|
| `BAD_REQUEST` | 400 | Entrada inválida pelo zod; `data.issues` traz as mensagens |
| `UNPROCESSABLE_CONTENT` | 422 | Regra de negócio violada, como moedas diferentes ou pedido sem itens |
| `PRECONDITION_FAILED` | 412 | Pedido já realizado |
| `NOT_FOUND` | 404 | Pedido que não existe |

A tela mostra a mensagem em português de cada caso, por `apps/web/src/lib/error-message.ts`.

## Persistência

Em memória: os dados somem quando o Next.js para. Para usar PostgreSQL, gere o projeto com `--with postgres`.
