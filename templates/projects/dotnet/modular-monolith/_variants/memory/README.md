# SampleApp

API .NET como monólito modular com DDD, gerada pelo `romanctl`.

```bash
romanctl run sample-app
dotnet test
```

Endereço: http://localhost:5083. Documento OpenAPI: `/openapi/v1.json`.

## Módulos

Um só processo e um só deploy, com o código dividido em módulos que não enxergam o interior uns dos outros.
Cada módulo tem dois projetos:

| Projeto | Conteúdo |
|---|---|
| `SampleApp.<Módulo>` | A implementação, toda `internal`: `Domain`, `Application`, `Infrastructure`, `Endpoints` e o registro em `<Módulo>Module.cs` |
| `SampleApp.<Módulo>.Contracts` | O que os outros módulos podem usar: interfaces de consulta e eventos de integração |

| Módulo | Rotas | Esquema |
|---|---|---|
| `Catalog` | `POST /catalog/products`, `GET /catalog/products/{id}`, `GET /catalog/products?limit=` | `catalog` |
| `Ordering` | `POST /orders`, `POST /orders/{id}/items`, `POST /orders/{id}/place`, `GET /orders/{id}`, `GET /orders?limit=` | `ordering` |

`SampleApp.Shared` guarda o mínimo que todos usam: `Money`, `DomainException`, o barramento de eventos em
processo, o tratamento de erros como ProblemDetails e o registro de um `DbContext` por módulo.
`SampleApp.Api` só compõe os módulos, em `Modules.cs`.

## Como os módulos conversam

- **Consulta síncrona por contrato.** Para adicionar um item, o `Ordering` pede o produto ao `ICatalogApi`, de
  `SampleApp.Catalog.Contracts`, e recebe um `ProductSnapshot` com nome e preço. Ele nunca toca no `Product`
  nem nos dados do catálogo.
- **Evento de integração assíncrono.** Ao realizar o pedido, o `Ordering` publica `OrderPlaced`, de
  `SampleApp.Ordering.Contracts`, no `IEventBus`. O evento entra numa fila em memória e é entregue fora da
  requisição, num escopo próprio, a cada `IIntegrationEventHandler<OrderPlaced>`. O `Catalog` trata o evento e
  soma a quantidade vendida de cada produto (`soldQuantity`).

A fila é em memória: um evento publicado e ainda não tratado se perde se o processo parar.

Os testes de arquitetura, com NetArchTest, garantem que um módulo só usa os contratos dos outros, que contratos
não dependem de implementação nem de frameworks, que o `Shared` não conhece nenhum módulo e que o domínio de
cada módulo não depende das outras camadas.

As regras ficam no domínio; os endpoints só traduzem HTTP. Corpo inválido devolve 400, regra violada devolve 422
e recurso inexistente devolve 404, sempre como ProblemDetails.

## Persistência

Em memória, com o provedor InMemory do EF Core e um `DbContext` por módulo: os dados somem quando a API para.
Para usar PostgreSQL, com um esquema por módulo, gere o projeto com `--with postgres`.

## Gerar código

```bash
romanctl add module Shipment                  # módulo Shipments com a entidade Shipment: contratos, rotas e testes
romanctl add resource Category --in Catalog   # entidade com CRUD dentro de um módulo existente
```
