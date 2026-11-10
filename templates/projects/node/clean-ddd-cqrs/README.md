# SampleApp

API Node com NestJS, Clean Architecture, DDD e CQRS, em TypeScript, gerada pelo `romanctl`.

```bash
romanctl run sample-app
npm test
```

Endereço: http://localhost:3000. A porta muda com a variável `PORT`.

| Script | O que faz |
|---|---|
| `npm start` | Compila para `dist/` e sobe a API, lendo o `.env` se existir |
| `npm test` | Confere os tipos e roda os testes com Vitest |
| `npm run build` | Só compila |

## Camadas

As dependências apontam sempre para dentro. O teste em `test/architecture.test.ts` garante isso.

| Pasta | Conteúdo | Depende de |
|---|---|---|
| `src/domain` | Agregados, objetos de valor, eventos e regras de negócio | nada, nem pacotes |
| `src/application` | Comandos, consultas, handlers e as portas que eles precisam (`OrderReadStore`, `Clock`) | domain, `@nestjs/cqrs` |
| `src/infrastructure` | Persistência e relógio | application, domain |
| `src/api` | Controllers, tratamento de erros e o módulo que monta tudo | todas |

## CQRS

Os controllers só montam a mensagem e a enviam pelo `CommandBus` ou pelo `QueryBus` do `@nestjs/cqrs`. Comandos
alteram estado pelo repositório do agregado; consultas leem pelo `OrderReadStore` e devolvem views. Os eventos que o
agregado levanta, como `OrderPlaced`, saem pelo `EventBus` depois de salvos.

## Endpoints

| Método | Caminho | O que faz |
|---|---|---|
| `POST` | `/orders` | Cria um pedido: `{"customerName": "Ana"}` |
| `POST` | `/orders/{id}/items` | Adiciona um item: `{"productName": "Caneta", "quantity": 2, "unitPrice": 3.5}` |
| `POST` | `/orders/{id}/place` | Realiza o pedido |
| `GET` | `/orders/{id}` | Um pedido |
| `GET` | `/orders?limit=50` | Os pedidos mais recentes |
| `GET` | `/health` | Saúde da API |

Regras de negócio violadas respondem 422, pedidos que não existem, 404, e corpos inválidos, 400, sempre como
`application/problem+json`.

## Persistência

PostgreSQL com `pg`, lido de `DATABASE_URL`. O `romanctl up sample-app` cria o banco `sample_app` e grava a variável
no `.env`. As tabelas são criadas na inicialização com `CREATE TABLE IF NOT EXISTS`, que não evolui o esquema: adote
migrações antes de ter dados que importam.

Os testes de HTTP usam o repositório em memória. O teste do repositório PostgreSQL só roda com `DATABASE_URL`.
