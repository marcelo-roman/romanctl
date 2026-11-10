# Persistência

PostgreSQL com pgx, lido de `DATABASE_URL`. O `romanctl up sample-app` cria o banco `sample_app` e grava a variável no
`.env`. Os dois serviços usam o mesmo banco, cada um no seu esquema, e nenhum lê o esquema do outro:

| Esquema | Tabelas |
|---|---|
| `orders` | `orders`, `order_items` e `outbox` |
| `billing` | `invoices` (uma por pedido) e `inbox` (os ids de eventos já processados) |

Cada serviço cria o próprio esquema na inicialização, sob `pg_advisory_xact_lock`, para os dois poderem subir ao
mesmo tempo. `CREATE TABLE IF NOT EXISTS` não evolui o esquema: adote migrações antes de ter dados que importam.

Os testes em `internal/services/*/adapters/postgres` e `internal/platform/postgres` rodam contra o banco de
`DATABASE_URL` e são pulados sem ela.
