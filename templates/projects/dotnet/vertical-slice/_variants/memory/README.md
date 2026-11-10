# SampleApp

API .NET com Vertical Slice Architecture e Minimal APIs, gerada pelo `romanctl`.

```bash
romanctl run sample-app
dotnet test
```

Endereço: http://localhost:5081. Documento OpenAPI: `/openapi/v1.json`.

## Fatias

Cada caso de uso é um arquivo em `Features/<Funcionalidade>/`, com tudo o que ele precisa: os records de
requisição e resposta, a validação do corpo, o handler e o mapeamento do endpoint (`MapEndpoint`). Não há
mediador nem camada de serviço compartilhada. Uma fatia não usa outra; um teste em `ArchitectureTests` garante isso.

| Pasta | Conteúdo |
|---|---|
| `Features/Orders` | `CreateOrder`, `AddItem`, `PlaceOrder`, `GetOrder` e `ListOrders` |
| `Domain` | O agregado `Order`, o objeto de valor `Money` e as regras de negócio |
| `Data` | `AppDbContext` e o mapeamento do EF Core |
| `Endpoints.cs` | A lista das fatias registradas |

As regras ficam no domínio; as fatias só orquestram. Corpo inválido devolve 400, regra violada devolve 422 e
recurso inexistente devolve 404, sempre como ProblemDetails.

## Persistência

Em memória, com o provedor InMemory do EF Core: os dados somem quando a API para. Para usar PostgreSQL, gere o
projeto com `--with postgres`.
