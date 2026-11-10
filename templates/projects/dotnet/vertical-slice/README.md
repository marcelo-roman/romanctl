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

PostgreSQL com EF Core. Rode `romanctl start postgres` antes de subir a API.
Em desenvolvimento, o banco `sample_app` é criado na inicialização com `EnsureCreated`, que não
serve para evoluir o esquema: adote migrações do EF Core antes de ter dados que importam.

Os testes das fatias usam o provedor em memória do EF Core. O teste contra o PostgreSQL só roda com a variável
`ConnectionStrings__Default` definida.
