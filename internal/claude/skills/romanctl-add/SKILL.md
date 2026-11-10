---
name: romanctl-add
description: Gera código dentro de um projeto pelos geradores do romanctl, no padrão do template (resource, command, query, event, value-object, controller, repository, test no .NET Clean/DDD/CQRS; page, component, hook, api no React e Next.js; route, middleware, service no Express; handler, model, service no Go; resource nos templates de Go, Java, Python, Node e .NET vertical slice; resource e event no .NET event sourcing). Use quando o usuário pedir uma entidade, um CRUD, um endpoint, uma página ou um componente num projeto criado pelo romanctl.
---

# Geradores de código pelo romanctl

Siga as regras gerais da skill `romanctl`. Gere pelo romanctl primeiro e só depois edite o
que foi gerado: assim o código segue as convenções do template (pastas, nomes, registro em
DI, rotas).

## O que dá para gerar neste projeto

```bash
romanctl add --list --project <projeto> --json
```

Devolve `target` (o template do projeto) e `generators` (os tipos aceitos). Sem
`--project`, vale o projeto da pasta atual.

## Gerar

```bash
romanctl add <tipo> <Nome> --project <projeto> [--dry-run] [--json]
romanctl add resource Product --project loja                 # .NET: entidade + CRUD completo
romanctl add command ArchiveProduct --in Product --project loja
romanctl add query ListActiveProducts --in Product --project loja
romanctl add page sobre --project site                       # Next.js/React
romanctl add route pedidos --project api                     # Express
```

- `--in <Agregado>` é obrigatório para `command`, `query`, `event` e `value-object`.
- `--dry-run` lista os arquivos antes de criar; use para mostrar o plano ao usuário quando a
  mudança for grande.
- Arquivo existente não é sobrescrito; `--force` sobrescreve, então só com pedido explícito.
- Depois de gerar, compile e rode os testes do projeto com as ferramentas dele (`dotnet build`,
  `npm run build`, `go build ./...`) antes de dizer que terminou.

Se o tipo pedido não existe em `generators`, escreva o código à mão seguindo o padrão dos
arquivos vizinhos e diga ao usuário que não havia gerador.
