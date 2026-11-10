# SampleApp

Dois microsserviços Go, pedidos (`orders`) e faturamento (`billing`), gerados pelo `romanctl`. Cada um sobe sozinho e
pode ser publicado sozinho; eles só se falam por eventos de integração.

```bash
romanctl run sample-app        # sobe os dois, com [orders] e [billing] no começo de cada linha do log
go test ./...
```

| Serviço | Endereço | Programa |
|---|---|---|
| `orders` | http://localhost:8091 | `go run ./cmd/orders` |
| `billing` | http://localhost:8092 | `go run ./cmd/billing` |

A porta de cada um muda com a variável `PORT` no ambiente daquele processo, como no `docker-compose.yml`. Não a coloque
no `.env`: ele vale para os dois processos, e eles disputariam a mesma porta.

## Estrutura

| Pasta | Conteúdo |
|---|---|
| `cmd/orders`, `cmd/billing` | Monta os adaptadores de um serviço e sobe o servidor |
| `internal/services/orders` | Pedidos: `domain`, `application` e `adapters` (HTTP, memória, PostgreSQL) |
| `internal/services/billing` | Faturamento: `domain`, `application` e `adapters` |
| `internal/platform` | O pouco que os dois compartilham: o envelope dos eventos, o relay do outbox, os transportes e as funções de HTTP |
| `internal/e2e` | O fluxo inteiro num processo só: pedido, outbox, relay, broker em memória e fatura |

Um serviço nunca importa o outro, o domínio não importa adaptadores e a plataforma não conhece os serviços. O teste em
`internal/architecture` garante as três regras. O contrato entre eles é o JSON do evento: o faturamento lê só os
campos de que precisa.

## Fluxo

1. `POST /orders/{id}/place` realiza o pedido e grava o evento `orders.order-placed` no outbox, na mesma transação
   do pedido.
2. O relay do outbox, dentro do `orders`, publica os eventos pendentes e só marca como enviado depois que o
   transporte confirma. Se falhar, tenta de novo com espera crescente: a entrega é pelo menos uma vez.
3. O `billing` recebe o evento e emite a fatura. O id do evento fica registrado na mesma operação que grava a fatura,
   então um evento repetido não gera outra.

## Endpoints

| Serviço | Método | Caminho | O que faz |
|---|---|---|---|
| orders | `POST` | `/orders` | Cria um pedido: `{"customerName": "Ana"}` |
| orders | `POST` | `/orders/{id}/items` | Adiciona um item: `{"productName": "Caneta", "quantity": 2, "unitPrice": 3.5}` |
| orders | `POST` | `/orders/{id}/place` | Realiza o pedido |
| orders | `GET` | `/orders/{id}` | Um pedido |
| orders | `GET` | `/orders?limit=50` | Os pedidos mais recentes |
| billing | `GET` | `/invoices/{orderId}` | A fatura de um pedido |
| billing | `GET` | `/invoices?limit=50` | As faturas mais recentes |
| ambos | `GET` | `/health` | Saúde do serviço |

JSON inválido responde 400, regra de negócio violada 422 e o que não existe 404, sempre como `application/problem+json`.

## Mensageria e persistência

- [docs/messaging.md](docs/messaging.md): como os eventos viajam de um serviço para o outro.
- [docs/persistence.md](docs/persistence.md): onde ficam os dados, o outbox e a inbox.
