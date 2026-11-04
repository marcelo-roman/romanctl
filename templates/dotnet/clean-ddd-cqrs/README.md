# {{.Name}}

API .NET com Clean Architecture, DDD e CQRS, gerada pelo `romanctl`.

```bash
romanctl run {{.Lower}}
dotnet test
```

Endereço: http://localhost:5080. Documento OpenAPI: `/openapi/v1.json`.

## Camadas

As dependências apontam sempre para dentro. Um teste em `ArchitectureTests` garante isso.

| Projeto | Conteúdo | Depende de |
|---|---|---|
| `{{.Name}}.Domain` | Agregados, objetos de valor, eventos e regras de negócio | nada |
| `{{.Name}}.Application` | Comandos, consultas e handlers | Domain |
| `{{.Name}}.Infrastructure` | Persistência | Application |
| `{{.Name}}.Api` | Endpoints HTTP | Application, Infrastructure |

## CQRS

Comandos alteram estado pelo repositório do agregado. Consultas leem pelo `IOrderReadStore` e
devolvem DTOs. O `IDispatcher` encontra o handler de cada mensagem pela injeção de dependência.

## Persistência

{{if .Postgres -}}
PostgreSQL com EF Core. Rode `romanctl start postgres` antes de subir a API.
Em desenvolvimento, o banco `{{.DbName}}` é criado na inicialização com `EnsureCreated`, que não
serve para evoluir o esquema: adote migrações do EF Core antes de ter dados que importam.
{{- else -}}
Em memória: os dados somem quando a API para. Para usar PostgreSQL, gere o projeto com
`--with postgres`.
{{- end}}
