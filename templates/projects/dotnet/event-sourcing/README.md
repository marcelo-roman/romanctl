# SampleApp

API .NET com event sourcing, Clean Architecture, DDD e CQRS, gerada pelo `romanctl`.

```bash
romanctl run sample-app
dotnet test
```

Endereço: http://localhost:5082. Documento OpenAPI: `/openapi/v1.json`.

## Camadas

As dependências apontam sempre para dentro. Um teste em `ArchitectureTests` garante isso.

| Projeto | Conteúdo | Depende de |
|---|---|---|
| `SampleApp.Domain` | Agregados, eventos e regras de negócio | nada |
| `SampleApp.Application` | Comandos, consultas, projeções e as portas `IEventStore` e `IReadModelStore` | Domain |
| `SampleApp.Infrastructure` | Event store e modelos de leitura | Application |
| `SampleApp.Api` | Endpoints HTTP | Application, Infrastructure |

## Event sourcing

O estado de um pedido é a sequência dos seus eventos: `OrderCreated`, `ItemAdded` e `OrderPlaced`. O agregado
decide levantando eventos e se reconstrói aplicando cada um no `When`. Gravar acrescenta os eventos novos ao fluxo,
conferindo a versão que foi lida: se outra operação gravou antes, a API responde 409.

As consultas não reconstroem o agregado. Elas leem o `OrderSummary`, um modelo de leitura que a
`OrderSummaryProjection` atualiza a cada evento gravado. `GET /orders/{id}/events` mostra o fluxo.

| Método | Caminho | |
|---|---|---|
| `POST` | `/orders` | cria um pedido |
| `POST` | `/orders/{id}/items` | adiciona um item |
| `POST` | `/orders/{id}/place` | realiza o pedido |
| `GET` | `/orders/{id}` | o resumo do pedido |
| `GET` | `/orders?limit=50` | os pedidos mais recentes |
| `GET` | `/orders/{id}/events` | os eventos do pedido |

## Persistência

PostgreSQL com [Marten](https://martendb.io): os eventos vão para o event store do Marten, e as projeções rodam na
mesma transação. As tabelas são criadas na primeira gravação. Rode `romanctl start postgres` antes de subir a API.

Os testes em `Postgres/` rodam contra o banco quando `ConnectionStrings__Default` está definida e são pulados sem ela.
Os testes HTTP usam o event store em memória.
