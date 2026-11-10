# SampleApp

API Rust com arquitetura hexagonal e DDD, gerada pelo `romanctl`.

```bash
romanctl run sample-app
cargo test
cargo clippy --all-targets -- -D warnings
```

Endereço: http://localhost:3100. A porta muda com a variável `PORT`.

## Camadas

As dependências apontam sempre para dentro. O teste em `tests/architecture.rs` garante isso.

| Pasta | Conteúdo | Depende de |
|---|---|---|
| `src/domain` | Agregados, objetos de valor, eventos e regras de negócio | `chrono`, `thiserror`, `uuid` |
| `src/application` | Casos de uso e as portas que eles precisam (`OrderRepository`, `EventPublisher`) | domain |
| `src/adapters` | Implementações das portas: HTTP com axum, persistência e eventos | application, domain |
| `src/bootstrap.rs` | Monta os adaptadores em cada caso de uso | todos |
| `src/main.rs` | Lê o ambiente e sobe o servidor | bootstrap |

## Endpoints

| Método | Caminho | O que faz |
|---|---|---|
| `POST` | `/orders` | Cria um pedido: `{"customerName": "Ana"}` |
| `POST` | `/orders/{id}/items` | Adiciona um item: `{"productName": "Caneta", "quantity": 2, "unitPrice": 3.5}` |
| `POST` | `/orders/{id}/place` | Realiza o pedido |
| `GET` | `/orders/{id}` | Um pedido |
| `GET` | `/orders?limit=50` | Os pedidos mais recentes, até 200 |
| `GET` | `/health` | Saúde da API |

Regras de negócio violadas respondem 422, pedidos que não existem, 404, e corpos inválidos, 400.

## Persistência

PostgreSQL com sqlx, lido de `DATABASE_URL`. O `romanctl up sample-app` cria o banco `sample_app` e grava a variável no
`.env`. As tabelas são criadas na inicialização com `CREATE TABLE IF NOT EXISTS`, sob um `pg_advisory_xact_lock` para
que duas instâncias subindo juntas não disputem o esquema. Isso não evolui o esquema: adote migrações antes de ter dados
que importam.

O teste em `tests/postgres_orders.rs` só roda com `DATABASE_URL` definida; sem ela, ele avisa e passa.
