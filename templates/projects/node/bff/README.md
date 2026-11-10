# SampleApp

Monorepo Node com o padrão Backend for Frontend, gerado pelo `romanctl`: uma API de pedidos, um BFF e uma web em
React, cada um em seu programa.

```bash
romanctl up sample-app
npm test
```

| Programa | Pasta | Endereço | O que é |
|---|---|---|---|
| `web` | `apps/web` | http://localhost:3300 | SPA em React com Vite; fala só com o BFF |
| `bff` | `apps/bff` | http://localhost:3301 | Fastify; monta as telas da web a partir da API |
| `api` | `apps/api` | http://localhost:3302 | Fastify; pedidos e catálogo, com as regras de negócio |

O `romanctl run` e o `romanctl up` sobem os três juntos, com `[api]`, `[bff]` e `[web]` no começo de cada linha do
log. Sem o `romanctl`, `npm run dev` na raiz faz o mesmo.

| Script na raiz | O que faz |
|---|---|
| `npm run dev` | Sobe a API, o BFF e a web |
| `npm test` | Confere os tipos e roda os testes de todos os programas e o teste de arquitetura |
| `npm run build` | Compila a API e o BFF para `dist/` e gera a web estática em `apps/web/dist` |

| Variável | Quem lê | Padrão |
|---|---|---|
| `API_PORT` | api | `3302` |
| `DATABASE_URL` | api | `postgres://postgres:postgres@127.0.0.1:5432/sample_app` |
| `BFF_PORT` | bff | `3301` |
| `API_URL` | bff | `http://localhost:3302` |
| `WEB_DIST` | bff | nenhum; com uma pasta, o BFF também serve a web compilada |
| `BFF_URL` | web, no `vite` | `http://localhost:3301` |

## Backend for Frontend

A web não conhece a API. Ela pede telas prontas ao BFF, que chama a API, junta e formata o que a tela precisa:

| Método | Caminho no BFF | O que faz |
|---|---|---|
| `GET` | `/screens/orders` | A lista de pedidos, com a quantidade de itens e o total formatado (`R$ 19,00`) |
| `GET` | `/screens/orders/{id}` | O pedido e as sugestões de produtos do catálogo, numa resposta só |
| `POST` | `/actions/orders` | Cria um pedido: `{"customerName": "Ana"}` |
| `POST` | `/actions/orders/{id}/items` | Adiciona um produto do catálogo: `{"productId": "...", "quantity": 2}` |
| `POST` | `/actions/orders/{id}/place` | Realiza o pedido |
| `GET` | `/health` | Saúde do BFF |

O BFF não tem regra de negócio. Ele busca nome e preço do produto no catálogo antes de adicionar o item, troca os
erros da API por mensagens para a tela (`{"message": "..."}`, com 400, 404, 422 ou 502 quando a API não responde) e
repassa o que a API decide, como `canAddItems` e `canPlace`. Na web, o Vite encaminha `/screens` e `/actions` para o
BFF, então a página e o BFF ficam na mesma origem, como em produção.

## API

| Método | Caminho | O que faz |
|---|---|---|
| `POST` | `/orders` | Cria um pedido: `{"customerName": "Ana"}` |
| `POST` | `/orders/{id}/items` | Adiciona um item: `{"productName": "Caneta", "quantity": 2, "unitPrice": 3.5, "currency": "BRL"}` |
| `POST` | `/orders/{id}/place` | Realiza o pedido |
| `GET` | `/orders/{id}` | Um pedido |
| `GET` | `/orders?limit=50` | Os pedidos mais recentes |
| `POST` | `/catalog/products` | Cadastra um produto: `{"name": "Caneta", "price": 3.5}` |
| `GET` | `/catalog/products` | O catálogo, em ordem de nome |
| `GET` | `/catalog/products/{id}` | Um produto |
| `GET` | `/health` | Saúde da API |

Regras: o pedido nasce como rascunho e precisa do nome do cliente; a quantidade é um inteiro maior que zero; todos os
itens usam a mesma moeda; só um rascunho com ao menos um item pode ser realizado, e depois disso não muda mais.
Regras violadas respondem 422, pedidos e produtos que não existem, 404, e corpos inválidos, 400, sempre como
`application/problem+json`. Na primeira subida, a API cadastra três produtos de exemplo se o catálogo estiver vazio.

| Pasta | Conteúdo | Depende de |
|---|---|---|
| `src/domain` | `Order`, `OrderItem`, `Money` e `CatalogProduct`, com as regras | nada |
| `src/repositories` | As interfaces e as implementações em `memory/` e `postgres/` | domain |
| `src/services` | Os casos de uso, que carregam, mudam e salvam | domain e as interfaces dos repositórios |
| `src/routes` | Rotas do Fastify, schemas e a tradução de erros para problem JSON | services, domain |

## Arquitetura

`test/architecture.test.ts`, na raiz, garante que os programas só conversam por HTTP: a web não importa nada da API
nem do BFF e só chama `/screens` e `/actions`, o BFF não importa nada da API e a API não sabe que o BFF e a web
existem. Cada programa tem os próprios tipos do contrato que consome. Na API, `test/layering.test.ts` garante a
direção das camadas.

## Persistência

PostgreSQL com `pg`, lido de `DATABASE_URL`. O `romanctl up sample-app` cria o banco `sample_app` e grava a variável
no `.env` da raiz, que a API lê. As tabelas são criadas na inicialização com `CREATE TABLE IF NOT EXISTS`, sob
`pg_advisory_xact_lock`, que não evolui o esquema: adote migrações antes de ter dados que importam.

Os testes de HTTP usam os repositórios em memória. O teste dos repositórios PostgreSQL só roda com `DATABASE_URL`.

## Testes

| Programa | O que os testes cobrem |
|---|---|
| api | Domínio, serviços, HTTP com `inject` e, com `DATABASE_URL`, os repositórios PostgreSQL |
| bff | Telas e ações contra uma API falsa num servidor HTTP de verdade: a junção, a resolução do produto e a tradução de erros |
| web | Páginas com Testing Library e jsdom, com o `fetch` simulado |

## Publicar

O `romanctl publish` gera um `Dockerfile` só, com o argumento `APP` (`api` ou `bff`), e um `compose.yaml` com os dois
serviços. A imagem do BFF leva a web compilada e a serve na mesma porta, em `WEB_DIST`, então em produção são dois
containers: a API, que só o BFF enxerga, e o BFF, publicado na porta 3301.

```bash
DATABASE_URL=postgres://usuario:senha@banco:5432/sample_app docker compose up --build
```
