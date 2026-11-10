# Sample App

Front-end React para a API de pedidos, organizado em Feature-Sliced Design, gerado pelo `romanctl`.

```bash
romanctl run sample-app
npm test
```

Endereço: http://localhost:5173.

| Script | O que faz |
|---|---|
| `npm run dev` | Sobe o Vite na porta 5173, com o proxy para a API |
| `npm test` | Confere os tipos, roda o ESLint com as regras de camadas e os testes com Vitest |
| `npm run build` | Confere os tipos e gera `dist/` |
| `npm run lint` | Só o ESLint |

## A API

O front fala com a API de pedidos que todos os templates de back-end do `romanctl` expõem, com o mesmo contrato:

| Método | Caminho | O que faz |
|---|---|---|
| `POST` | `/orders` | Cria um pedido: `{"customerName": "Ana"}` → `201 {"id": "..."}` |
| `POST` | `/orders/{id}/items` | Adiciona um item: `{"productName": "Caneta", "quantity": 2, "unitPrice": 3.5, "currency": "BRL"}` → `204` |
| `POST` | `/orders/{id}/place` | Realiza o pedido → `204` |
| `GET` | `/orders/{id}` | Um pedido, com itens, total e moeda |
| `GET` | `/orders?limit=50` | Os pedidos mais recentes |

Erros chegam como problem JSON (`{"status", "title", "detail"}`) com 400, 404 ou 422, e a tela mostra o `detail`.

### Para onde aponta

Em desenvolvimento, o navegador chama `/api/...` no próprio Vite, e o Vite repassa para `VITE_API_URL` sem o
prefixo `/api`. Assim a API não precisa liberar CORS. Sem `VITE_API_URL`, o destino é `http://localhost:8090`, a
porta do template Go. Para outro back-end, crie um `.env` a partir do `.env.example`:

| Back-end | `VITE_API_URL` |
|---|---|
| Go hexagonal e DDD | `http://localhost:8090` |
| .NET Clean, DDD e CQRS | `http://localhost:5080` |
| .NET event sourcing | `http://localhost:5082` |
| .NET monólito modular | `http://localhost:5083` |
| Node Clean, DDD e CQRS | `http://localhost:3000` |
| Python hexagonal e DDD | `http://localhost:8000` |
| Java hexagonal e DDD | `http://localhost:8080` |
| Kotlin hexagonal e DDD | `http://localhost:8081` |
| PHP hexagonal e DDD | `http://localhost:8002` |
| Rust hexagonal e DDD | `http://localhost:3100` |

Confira a porta no `romanctl.yaml` do back-end. Com `VITE_API_PROXY=false`, o navegador chama `VITE_API_URL`
direto, e aí a API precisa responder com CORS.

```bash
romanctl new pedidos-api --stack go --arch hexagonal,ddd
romanctl up pedidos-api -d
romanctl up sample-app -d
```

## Camadas

Cada camada só importa das que estão abaixo dela, e de outra fatia só pelo `index.ts`, a API pública da fatia. Duas
fatias da mesma camada não se importam. O ESLint, com `eslint-plugin-boundaries`, garante as três regras e roda no
`npm test`.

| Camada | Conteúdo | Pode importar |
|---|---|---|
| `src/app` | Entrada, provedores (TanStack Query), rotas, layout e estilos | todas |
| `src/pages` | `orders-list`, `order-details` e `not-found` | widgets, features, entities, shared |
| `src/widgets` | `orders-table` e `order-summary`, blocos prontos de tela | features, entities, shared |
| `src/features` | `create-order`, `add-order-item` e `place-order`, cada uma com `ui`, `model` e `api` | entities, shared |
| `src/entities` | `order`: tipos, mapeamento do contrato, formatação de dinheiro, consultas | shared |
| `src/shared` | Cliente HTTP com o erro de problem JSON, configuração, utilitários e o kit de UI | shared |

Os componentes só desenham. Formatação, conversão de formulário e regras ficam em `model` e `lib`, testados sem
React.

## Testes

Os testes ficam ao lado do código, em `*.test.ts(x)`. O MSW intercepta as chamadas e responde como a API de
verdade, com um back-end em memória em `tests/support/orders-api.ts`. `tests/support` também tem o `renderRoutes`,
que monta o TanStack Query e um roteador em memória.

## Adicionar ao projeto

```bash
romanctl add entity Customer        # src/entities/customer: modelo, API, consultas e testes
romanctl add feature archive-order  # src/features/archive-order: formulário com ui, model, api e teste
romanctl add page Customers         # src/pages/customers, com a rota e o link no menu
```
