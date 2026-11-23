---
name: romanctl-add
description: Generates code inside a project through romanctl's generators, following the template's conventions (resource, command, query, event, value-object, controller, repository, test in .NET Clean/DDD/CQRS; page, component, hook, api in React and Next.js; route, middleware, service in Express; handler, model, service in Go; resource in the architecture templates; module and resource --in in the .NET modular monolith; resource and event in .NET event sourcing). Use when the user asks for an entity, a CRUD, an endpoint, a page or a component in a project created by romanctl.
---

# Code generators through romanctl

Follow the general rules of the `romanctl` skill. Generate through romanctl first and only then edit
what was generated: that way the code follows the template's conventions (folders, names, DI
registration, routes).

## What can be generated in this project

```bash
romanctl add --list --project <project> --json
```

Returns `target` (the project's template) and `generators` (the accepted types). Without
`--project`, the project in the current folder is used.

## Generate

```bash
romanctl add <type> <Name> --project <project> [--dry-run] [--json]
romanctl add resource Product --project store                # .NET: entity + full CRUD
romanctl add command ArchiveProduct --in Product --project store
romanctl add query ListActiveProducts --in Product --project store
romanctl add page about --project site                       # Next.js/React
romanctl add route orders --project api                      # Express
```

- `--in <Aggregate>` is required for `command`, `query`, `event` and `value-object`.
- `--plural <Name>` fixes the plural when the English rule gets it wrong, as in `add resource Person --plural People`.
- `--dry-run` lists the files before creating them; use it to show the plan to the user when the
  change is large.
- Existing files are not overwritten; `--force` overwrites, so only on explicit request.
- After generating, build and run the project's tests with its own tools (`dotnet build`,
  `npm run build`, `go build ./...`) before saying you are done.

If the requested type does not exist in `generators`, write the code by hand following the pattern of the
neighboring files and tell the user there was no generator.
