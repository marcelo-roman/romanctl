---
name: romanctl-projects
description: Creates, lists and removes projects through romanctl, from templates (Node, Express, React, Next.js, .NET Web API/MVC/Blazor/Console, Go) and from architecture templates (hexagonal Go, Rust, Java, Kotlin, PHP and Python with DDD; .NET Clean/DDD/CQRS, event sourcing, vertical slice and modular monolith; Node NestJS Clean/DDD/CQRS and layered Fastify; layered Django; with Postgres). Use when the user asks for a new project, an API, a website, a web app, or wants to know which projects exist.
---

# Projects through romanctl

Follow the general rules of the `romanctl` skill.

## Before creating: what exists

```bash
romanctl types                # simple types: node, express, react, react-ts, next, dotnet-*, go
romanctl catalog --json       # architecture templates, valid combinations and generators
romanctl list --json          # projects that already exist
```

Choose the template from what the user described and say which one you chose. If nothing fits
(mobile, desktop, a language or architecture without a template...), follow "When romanctl does not cover the request" in the
`romanctl` skill.

## Create

Simple type: **the type comes before the name**.

```bash
romanctl new <type> <name>
romanctl new next site
romanctl new dotnet-webapi orders
romanctl new express api --no-git         # no git repository
romanctl new react dashboard --offline    # local templates and cache only
```

Architecture template: name, stack and architectures, and the services the project uses.

```bash
romanctl new <name> --stack <stack> --arch <a,b,c> [--with <services>] [--dry-run] [--json]
romanctl new store --stack dotnet --arch clean,ddd,cqrs --with postgres
```

- The valid `--arch` combinations and the services accepted by `--with` are in
  `romanctl catalog --json`; do not invent combinations.
- `--dry-run` shows the files that would be created, and `missingTools` in `--json` tells which tools
  are missing to run the project.
- After generating, `new` runs the template setup (`dotnet restore`, `npm install`...). Without the
  tools, or with `--offline`, the setup is deferred and `setup` in `--json` says why
  (`skipped`, `offline`, `failed` or `done`). If a tool is missing, install it through the `romanctl-tools` skill
  and run the setup in the project folder.
- Names: lowercase letters, digits, `-`, `_` or `.`; for a template, starting with a letter.
- Templates live outside the binary. If `catalog` comes back empty or a template that should exist is missing,
  run `romanctl templates update --json` (needs internet) and check with `romanctl templates --json`.
- The project is created in `~/projetos/<name>` (or in `$ROMANCTL_PROJETOS`).

After creating, the natural next step is to start the environment (`romanctl-dev` skill) and, if the
user wants, generate code (`romanctl-add` skill).

## List and locate

```bash
romanctl list --json                     # name, type, folder, date
romanctl list --filter <text> --plain    # filtered, as text
```

Use the folder that `list --json` returns to read and edit the project's code.

## Remove

Deletes **all of the project's files**. Only on the user's explicit request for that
project, and stop whatever is running first:

```bash
romanctl down <name> --json
romanctl rm <name> --yes
```

## Open for the user

Only when they ask: `romanctl code <name>` opens it in code-server, and `romanctl open <name> --app`
opens the running app in the browser.
