# SampleApp

API Go com arquitetura hexagonal e DDD, gerada pelo `romanctl`.

```bash
romanctl run sample-app
go test ./...
```

Endereço: http://localhost:8090. A porta muda com a variável `PORT`.

## Camadas

As dependências apontam sempre para dentro. O teste em `internal/architecture` garante isso.

| Pasta | Conteúdo | Depende de |
|---|---|---|
| `internal/domain` | Agregados, objetos de valor, eventos e regras de negócio | só a biblioteca padrão |
| `internal/application` | Casos de uso e as portas que eles precisam (`Repository`, `EventPublisher`) | domain |
| `internal/adapters` | Implementações das portas: HTTP, persistência e eventos | application, domain |
| `cmd/api` | Monta os adaptadores e sobe o servidor | todos |

## Endpoints

| Método | Caminho | O que faz |
|---|---|---|
| `POST` | `/orders` | Cria um pedido: `{"customerName": "Ana"}` |
| `POST` | `/orders/{id}/items` | Adiciona um item: `{"productName": "Caneta", "quantity": 2, "unitPrice": 3.5}` |
| `POST` | `/orders/{id}/place` | Realiza o pedido |
| `GET` | `/orders/{id}` | Um pedido |
| `GET` | `/orders?limit=50` | Os pedidos mais recentes |
| `GET` | `/health` | Saúde da API |

Regras de negócio violadas respondem 422, e pedidos que não existem, 404.

## Persistência

PostgreSQL com pgx, lido de `DATABASE_URL`. O `romanctl up sample-app` cria o banco `sample_app` e grava a variável no
`.env`. As tabelas são criadas na inicialização com `CREATE TABLE IF NOT EXISTS`, que não evolui o esquema: adote
migrações antes de ter dados que importam.

