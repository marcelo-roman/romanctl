# SampleApp

API .NET com Clean Architecture, DDD e CQRS, gerada pelo `romanctl`.

```bash
romanctl run sample-app
dotnet test
```

Endereço: http://localhost:5080. Documento OpenAPI: `/openapi/v1.json`.

## Camadas

As dependências apontam sempre para dentro. Um teste em `ArchitectureTests` garante isso.

| Projeto | Conteúdo | Depende de |
|---|---|---|
| `SampleApp.Domain` | Agregados, objetos de valor, eventos e regras de negócio | nada |
| `SampleApp.Application` | Comandos, consultas e handlers | Domain |
| `SampleApp.Infrastructure` | Persistência | Application |
| `SampleApp.Api` | Endpoints HTTP | Application, Infrastructure |

## CQRS

Comandos alteram estado pelo repositório do agregado. Consultas leem pelo `IOrderReadStore` e
devolvem DTOs. O `IDispatcher` encontra o handler de cada mensagem pela injeção de dependência.

## Persistência

PostgreSQL com EF Core. Rode `romanctl start postgres` antes de subir a API.
Em desenvolvimento, o banco `sample_app` é criado na inicialização com `EnsureCreated`, que não
serve para evoluir o esquema: adote migrações do EF Core antes de ter dados que importam.
