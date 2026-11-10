---
name: romanctl-projects
description: Cria, lista e remove projetos pelo romanctl, a partir de templates (Node, Express, React, Next.js, .NET Web API/MVC/Blazor/Console, Go) e de templates de arquitetura (Go, Rust, Java, Kotlin, PHP e Python hexagonais com DDD; .NET Clean/DDD/CQRS, event sourcing, vertical slice e monólito modular; Node NestJS Clean/DDD/CQRS e Fastify em camadas; Django em camadas; com Postgres). Use quando o usuário pedir um projeto novo, uma API, um site, um app web, ou quiser saber que projetos existem.
---

# Projetos pelo romanctl

Siga as regras gerais da skill `romanctl`.

## Antes de criar: o que existe

```bash
romanctl types                # tipos simples: node, express, react, react-ts, next, dotnet-*, go
romanctl catalog --json       # templates de arquitetura, combinações válidas e geradores
romanctl list --json          # projetos que já existem
```

Escolha o template pelo que o usuário descreveu e diga qual escolheu. Se nada servir
(mobile, desktop, uma linguagem ou arquitetura sem template...), siga "Quando o romanctl não cobre o pedido" da skill
`romanctl`.

## Criar

Tipo simples: **o tipo vem antes do nome**.

```bash
romanctl new <tipo> <nome>
romanctl new next site
romanctl new dotnet-webapi pedidos
romanctl new express api --sem-git        # sem repositório git
romanctl new react painel --offline       # só templates e cache locais
```

Template de arquitetura: nome, stack e arquiteturas, e os serviços que o projeto usa.

```bash
romanctl new <nome> --stack <stack> --arch <a,b,c> [--with <serviços>] [--dry-run] [--json]
romanctl new loja --stack dotnet --arch clean,ddd,cqrs --with postgres
```

- As combinações válidas de `--arch` e os serviços aceitos em `--with` estão em
  `romanctl catalog --json`; não invente combinações.
- `--dry-run` mostra os arquivos que seriam criados, e `missingTools` no `--json` diz que ferramentas
  faltam para rodar o projeto.
- Depois de gerar, o `new` roda o preparo do template (`dotnet restore`, `npm install`...). Sem as
  ferramentas, ou com `--offline`, o preparo fica para depois e `setup` no `--json` diz por quê
  (`skipped`, `offline`, `failed` ou `done`). Faltando ferramenta, instale pela skill `romanctl-tools`
  e rode o preparo na pasta do projeto.
- Nomes: letras minúsculas, números, `-`, `_` ou `.`; num template, começando por letra.
- Os templates ficam fora do binário. Se `catalog` vier vazio ou faltar um template que deveria existir,
  rode `romanctl templates update --json` (precisa de internet) e confira com `romanctl templates --json`.
- O projeto nasce em `~/projetos/<nome>` (ou em `$ROMANCTL_PROJETOS`).

Depois de criar, o próximo passo natural é subir o ambiente (skill `romanctl-dev`) e, se o
usuário quiser, gerar código (skill `romanctl-add`).

## Listar e localizar

```bash
romanctl list --json                     # nome, tipo, pasta, data
romanctl list --filter <texto> --plain   # filtrado, em texto
```

Use a pasta que `list --json` devolve para ler e editar o código do projeto.

## Remover

Apaga **todos os arquivos** do projeto. Só com pedido explícito do usuário para aquele
projeto, e derrube antes o que estiver no ar:

```bash
romanctl down <nome> --json
romanctl rm <nome> --yes
```

## Abrir para o usuário

Só quando ele pedir: `romanctl code <nome>` abre no code-server, e `romanctl open <nome> --app`
abre o app rodando no navegador.
