# romanctl

[![CI](https://github.com/marcelo-roman/romanctl/actions/workflows/ci.yml/badge.svg)](https://github.com/marcelo-roman/romanctl/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/marcelo-roman/romanctl)](https://github.com/marcelo-roman/romanctl/releases/latest)

A CLI for developers. With a single command, it:

- **installs tools**: more than 170, such as kubectl, gh, the AWS and Azure CLIs, Terraform, Docker, SDKs for many
  languages and database clients;
- **creates projects**, including ready-made architectures (Clean Architecture, DDD and CQRS in .NET, hexagonal in Go,
  and more), and **generates code** inside them, such as full resources, controllers, commands and pages;
- **brings up the project environment**: PostgreSQL, MongoDB, Redis, MySQL, SQL Server, RabbitMQ and Kafka, the
  database, the `.env` and the app itself;
- **publishes to GitHub**, with CI and a Dockerfile;
- ships a **terminal dashboard**, in the spirit of k9s, to do all of that without memorizing commands.

It runs on Linux and macOS, on `amd64` and `arm64`, and also on WSL and on Termux with `proot-distro`. It speaks
English and Portuguese, following the system language.

![romanctl demo](docs/demo.gif)

The full walkthrough of every command is in [`docs/demo.mp4`](docs/demo.mp4). The web interface is shown in
[`docs/web-ui.mp4`](docs/web-ui.mp4). The terminal recordings are made with [vhs](https://github.com/charmbracelet/vhs) from the
repository root (`vhs docs/demo.tape`, `vhs docs/highlight.tape`); the web one with `docs/webrec`, which drives Chrome
against a running `romanctl web-ui`.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh | sh
```

That's it. Then run `romanctl` to see the command list, or `romanctl ui` to open the dashboard.

### What the installer does

1. Detects the system (Linux or macOS) and the architecture (`amd64` or `arm64`).
2. Downloads the binary from the [latest release](https://github.com/marcelo-roman/romanctl/releases/latest).
3. Checks the SHA-256 against the release's `SHA256SUMS` file. If it doesn't match, it stops without installing anything.
4. Installs to `~/.local/bin/romanctl`, or to `/usr/local/bin/romanctl` when run as root.
5. Warns you if that directory is not on your `PATH`.

Requirements: `curl` or `wget`, and `sha256sum` or `shasum`. They come preinstalled on most systems.

### Rather read the script first?

```bash
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh -o install.sh
less install.sh
sh install.sh
```

### Choosing the version or the directory

```bash
# a specific version
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh | ROMANCTL_VERSION=v0.1.0 sh

# another directory
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh | ROMANCTL_INSTALL_DIR="$HOME/bin" sh

# for every user on the machine
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh | sudo sh
```

### Adding it to the PATH

If the installer warns that `~/.local/bin` is not on your `PATH`, add it to your shell's file and open a new terminal:

| Shell | Command |
|---|---|
| bash | `echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc` |
| zsh | `echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc` |
| fish | `fish_add_path ~/.local/bin` |

### Checking

```bash
romanctl version      # romanctl 0.1.0
romanctl platform     # what it detected on your machine
romanctl doctor       # versions of the main tools
```

### Per system

**Linux.** Works on Debian, Ubuntu, Fedora, RHEL, Arch, openSUSE and Alpine. Tools are installed through the
distribution's package manager (`apt`, `dnf`, `pacman`, `zypper` or `apk`) or from the vendor's official binary.
Without root, only the package manager uses `sudo`. Services such as PostgreSQL and Kafka need Docker or Podman:
`romanctl tools install docker`.

**macOS.** Tools are installed through [Homebrew](https://brew.sh), which must be installed. For services,
`romanctl tools install docker` installs Colima with Docker.

**WSL.** Works like the Linux distribution you use.

**Termux (Android).** Use Ubuntu inside Termux and install it there:

```bash
pkg install proot-distro
proot-distro install ubuntu
proot-distro login ubuntu
apt update && apt install -y curl
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh | sh
```

Containers don't run under proot. `romanctl` detects that: PostgreSQL, MongoDB and Redis run as local processes, and
whatever needs a container is left out of the lists. Details in [Termux and proot](#termux-and-proot).

### From source

You need Go:

```bash
git clone https://github.com/marcelo-roman/romanctl.git
cd romanctl
go build -ldflags="-s -w" -o ~/.local/bin/romanctl ./cmd/romanctl
```

## Updating

```bash
romanctl self-update            # updates to the latest version
romanctl self-update --check    # only shows whether there is a new version
romanctl self-update --version v0.1.0
```

`self-update` checks the SHA-256 and runs the new version before replacing the current one: if it doesn't work on your
machine, nothing changes. If `romanctl` lives in a system directory such as `/usr/local/bin`, use `sudo`. Running the
installer again also updates it.

## Uninstalling

```bash
rm "$(command -v romanctl)"
```

That removes only the program. What it created stays until you delete it:

| Directory | Contents |
|---|---|
| `~/.local/share/romanctl` | local service data, logs, state and caches |
| `~/.config/romanctl` | chosen language, your own templates and generators |
| `~/projetos` | your projects |

Installed tools stay on the system. To remove them first, use `romanctl tools uninstall`. Service containers are named
`romanctl-<service>`, and their data lives in volumes with the same name.

## Getting started

**0. Log in.** romanctl only runs after logging in to Keycloak (realm `private` at `auth.marceloroman.com.br`):

```bash
romanctl login   # opens the browser; the session is stored in ~/.config/romanctl/session.json
```

**1. Look at the commands** and run whatever you want from the list:

```bash
romanctl         # command list: type to filter, Enter runs it and asks for whatever is missing
romanctl ui      # full-screen dashboard
```

In the dashboard, `Tab` switches screens, `/` filters, `Enter` shows what you can do with the item, and `?` shows
every key.

**2. Install the tools** you need:

```bash
romanctl tools list                  # what runs on your machine
romanctl tools install k8s           # pick from a list of the Kubernetes group
romanctl tools install gh node go    # or name them
```

**3. Create a project** and generate code in it:

```bash
romanctl new shop --stack dotnet --arch clean,ddd,cqrs --with postgres
cd ~/projetos/shop
romanctl add resource Product        # full CRUD, from the domain to the endpoints
```

**4. Bring everything up**: PostgreSQL, the database, the `.env` and the API.

```bash
romanctl up -d
curl http://localhost:5080/products
romanctl down
```

**5. Publish** to GitHub, as a private repository:

```bash
romanctl publish
```

Any command without arguments asks for what is missing. For help, use `romanctl <command> --help`.

## Commands

| Command | What for |
|---|---|
| `romanctl` | Lists the commands: type to filter, Enter runs it and asks for whatever is missing |
| `ui` | Opens the dashboard |
| `tools list`, `tools info`, `tools versions`, `tools install`, `tools uninstall`, `tools categories` | Tools |
| `new`, `add`, `catalog`, `types`, `templates` | Create projects and generate code |
| `up`, `down`, `run`, `logs` | Run projects with their environment |
| `code`, `open`, `list`, `rm` | Open, list and remove projects |
| `publish` | Publish to GitHub |
| `status`, `start`, `stop`, `restart`, `install`, `restore` | Services |
| `system`, `platform`, `doctor` | CPU, memory, battery and diagnostics |
| `offline` | Get ready to work without internet |
| `login`, `logout`, `whoami`, `iam` | Access and Keycloak administration |
| `web-ui` | Web interface on localhost |
| `claude` | Skills so Claude Code can operate the machine through romanctl |
| `config lang`, `self-update`, `menu` | Language, updates and the list menu |

Projects live in `~/projetos`; to use another directory, set `ROMANCTL_PROJETOS`. `--json` works on the query commands,
for scripts. Without a terminal, nothing is asked. Exit codes are 0 (success), 1 (execution failure) and 2 (incorrect
usage).

# Reference

## Login and Keycloak

```bash
romanctl login                   # authorization code with PKCE, redirect to 127.0.0.1
romanctl whoami                  # user, roles and when the session expires
romanctl logout                  # revokes the session in Keycloak and deletes the local one
```

Login happens in the `private` realm of `https://auth.marceloroman.com.br`, through a public client (`romanctl`) that
requires PKCE S256 and only accepts redirects to `http://127.0.0.1:<port>/callback`. The only way in is Google, with a
criarch Workspace account: romanctl sends `kc_idp_hint=google`, Keycloak goes straight to Google and rejects accounts
outside the domain. `private` only lets in people previously registered as operators
(`romanctl iam operator add <email>`). Roles come in the token as client roles of `iam-manager-api`. The token is
offline: the session lasts 30 days without use and is renewed on every command. Without internet, the saved session is
enough while it is valid; with internet, it is checked against Keycloak every 12 h, so a blocked user or a terminated
session is cut off right away. Without login, only `login`, `logout`, `whoami`, `help`, `version`, `self-update` and
`templates` run.

Pipelines have no browser: `ROMANCTL_CLIENT_ID` and `ROMANCTL_CLIENT_SECRET` (an IAM Manager API client in `private`,
with `client_credentials`) replace the login, without writing anything to disk. This repository's CI uses secrets with
the same names. `ROMANCTL_AUTH_ISSUER` and `ROMANCTL_IAM_API` (or `--issuer` and `--api` on login) point to another
Keycloak.

`romanctl iam` administers Keycloak through the [IAM Manager](https://iam-api.marceloroman.com.br/api/guide.md) API
with an **API client** (`client_credentials` in the `private` realm), created in IAM Manager → Operators → API clients.
The administered realms are `private` (operators) and `public` (application users):

```bash
romanctl iam credentials set --client-id iam-api-...   # asks for the secret without echo, checks and saves it
romanctl iam credentials                          # client, API and where they came from (never the secret)
romanctl iam me                                   # what the client is allowed to do
romanctl iam realms
romanctl iam users public --search ana
romanctl iam user create public ana --email ana@acme.com      # returns a link to set the password
romanctl iam apps public                                      # applications, roles and clients
romanctl iam app create public notes --roles reader,editor
romanctl iam user access public ana notes --roles reader      # access to an application
romanctl iam user disable public ana
romanctl iam operator add ana@criarch.com.br --global iam-auditor   # allows login to private
romanctl iam operator roles ana@criarch.com.br --realms public=admin
romanctl iam api-clients                                      # API clients
romanctl iam api-client create pipeline --realms public=admin # the secret is shown only here
romanctl iam audit
romanctl iam api GET /api/realms/public/settings   # any API route
```

The credential comes from `IAM_CLIENT_ID` and `IAM_CLIENT_SECRET` (with optional `IAM_API_URL` and `IAM_TOKEN_URL`) or
from `~/.config/iam-manager/credentials.env` (0600), the same file the `iam-manager` skill uses; the environment wins
over the file. The token lasts 5 minutes and renews itself, retrying once on 401. Permissions are the client's: global
roles `iam-admin` (everything), `iam-user-admin` (users, passwords and API keys), `iam-app-admin` (applications and
access), `iam-theme-admin` (themes) and `iam-auditor` (read-only), or the delegations `public:admin` and
`public:viewer`. Everything is audited with its `client_id`. Rotating your own client
(`romanctl iam api-client rotate`) writes the new secret to the file.

## Web interface

```bash
romanctl web-ui                  # http://localhost:7777
romanctl web-ui --port 8888 --open
```

[Watch it in action](docs/web-ui.mp4): create a project from a template, generate code, bring it up, read its logs,
browse services, tools and templates, follow jobs, switch language and remove the project.

A browser interface for what the CLI does: machine dashboard, projects (create from a template, bring up, take down,
logs, generate code, remove), services, tools (install, pick a version, remove), templates and running jobs, with live
output. Each action runs romanctl itself with `--json`, so the result is the same as on the command line. It follows
romanctl's language and the system's light or dark theme.

The server listens only on `127.0.0.1`, rejects requests with another host name and requires a token embedded in the
page, renewed on every run: another site open in the browser cannot call the API.

The code lives in [`web/`](web) (React, Vite and TypeScript) and the build goes to `internal/webui/dist`, embedded in
the binary. After changing the interface, run `npm run build` in `web/` and commit `dist`; CI checks that it is up to
date.

## Dashboard

`romanctl ui` opens a full-screen dashboard, in the spirit of k9s, with project, service, tool and system screens that
refresh by themselves.

At the top you see the machine, the version, how many projects and services are up, CPU and memory; beside them, the
keys that apply right now. The list sits in a frame with the screen name, the filter and the count in the title, each
row colored by its state, and a breadcrumb at the bottom. In a small terminal, the header shrinks to one line and the
keys move to the footer.

| Key | What it does |
|---|---|
| `Tab`, `1` to `4` | Switches screens |
| `:` | Jumps straight to `:tools`, `:svc`, `:prj`, `:sys`, or a category such as `:db` and `:k8s` |
| `/` | Filters the list |
| `Enter` | Shows the item's actions; in Tools, the available versions |
| `d`, `l` | Item details and logs; in Tools, details include the README, with `j`/`k` to scroll |
| `<` `>` | Switches category, in Tools |
| `?` | Shows every key of the current screen |
| `Esc`, `q` | Go back and quit |

The header shows only the keys that apply to the selected item. Long actions run in the background, with progress on
the row itself, and removing always asks for confirmation.

In Tools, when the cursor rests on a row, the dashboard looks up the version an install would bring and shows it in the
VERSION column, as in `1.29 → 1.31`. Installing and reinstalling ask for confirmation, with the version and the source.
`romanctl menu` opens the simpler list version.

## Language

English and Portuguese. The language follows the system: `LC_ALL`, `LC_MESSAGES`, `LANGUAGE`, `LANG` and, on macOS,
the account language. With no hint at all, it uses English. To choose:

```bash
romanctl config lang pt      # saves the choice; auto goes back to following the system
romanctl status --lang en    # only for this command
ROMANCTL_LANG=en romanctl status
```

On Termux with `proot-distro`, Android doesn't tell Ubuntu the language, so it's worth saving the choice. Projects
generated from the templates are in English.

The source code is in English. Messages are written in English in the code, and `internal/i18n/pt*.go` holds the
Portuguese translation; a test fails if a message is left untranslated.

## Tools

```bash
romanctl tools list                 # only what runs on this machine
romanctl tools list --all           # includes the rest, with the reason
romanctl tools info k9s             # installed, available, source, page and README
romanctl tools versions helm        # installable versions, newest first
romanctl tools install helm --version 3.18.4
romanctl tools install kubectl gh aws az docker
romanctl tools install k8s          # lists the group, with an All option
romanctl tools install k8s --yes    # installs the whole group
romanctl tools install k8s --dry-run
romanctl tools uninstall kubectl --yes
romanctl tools uninstall k8s        # lists what is installed in the group
```

The catalog lives in `internal/tools/catalog_*.go`, in groups: `sdk`, `node-tools`, `python-tools`, `build`, `git`,
`cloud`, `k8s`, `iac`, `containers`, `security`, `db`, `observability`, `network`, `shell` and `ai`. Downloads are
checked against the vendor's SHA-256 when one is published. After installing, the tool is executed to confirm it runs.

Without a terminal, or with `--json`, a group is installed whole, because there is nobody to answer the menu.

The tool list always shows the latest version of each one, cached for 6 hours so it doesn't hit the network every time.
In the dashboard, `Enter` on a tool opens the available versions and installs the chosen one; in `romanctl tools list`,
the option is "Choose another version". Versions come from the same source as the installer (GitHub releases,
HashiCorp, npm, PyPI, `apt-cache madison`, `dnf --showduplicates`, and the official Node, Go and Zig lists), without
pre-releases. With Homebrew and with the custom installers of aws, gcloud, dart, swift, rust and docker-cli, only the
latest exists.

`tools info` and the install show the version that will arrive, looked up in the same source the installer uses: GitHub
releases, `apt-cache policy` and equivalents, Homebrew, npm, PyPI or the official installer (dl.k8s.io, nodejs.org,
go.dev, HashiCorp). The README comes from GitHub, or from the PyPI description, and is rendered in the terminal; outside
a terminal, the raw markdown is printed.

The listing, the selection lists and the groups show only what runs on the machine, according to the system, the
architecture, the package manager, the libc and the detected capabilities: containers (namespaces and cgroups, and a VM
on macOS), virtualization (`/dev/kvm` or `kern.hv_support`) and a graphical interface (`DISPLAY` or
`WAYLAND_DISPLAY`). If detection gets it wrong, force it with `ROMANCTL_CONTAINERS`, `ROMANCTL_VM` and `ROMANCTL_GUI`,
set to `0` or `1`. What each tool requires is in `internal/tools/requirements.go`.

`uninstall` works the same way, showing only what is installed, and always asks for confirmation before removing.
`--yes` skips the list and the confirmation, and in that case `all` is refused. Before removing through `apt`, it
simulates the operation and refuses if it would take along an essential package or another explicitly installed one.

## Lists and progress

In a terminal, every list filters as you type, ignoring case and accents, and every word must appear in the row. The
first Esc clears the filter, and the second goes back. In multiple-choice lists, Space selects, and "All" or Ctrl+A
select only what the filter is showing.

In tool lists, ←/→ switch categories, and typing the category name also works. In commands, use `--category`, which
accepts names such as `db`, `databases` and `languages`, separated by commas. `romanctl tools categories` shows them
all.

`romanctl tools list` and `romanctl list` are navigable lists, with actions on each item. To print instead, use
`--plain` or `--filter <text>`; with a pipe, `--json` or no terminal, they print by themselves.

Whatever takes a while shows a progress animation, only in a terminal. Tool versions are cached until the binary
changes, and the size of the `offline` caches is kept for an hour.

## Architecture templates

```bash
romanctl new orders --stack dotnet --arch clean,ddd,cqrs --with postgres
romanctl new inventory --stack go --arch hexagonal,ddd --with postgres
```

| Stack | Architecture | Based on | Services |
|---|---|---|---|
| `dotnet` | `clean`, `ddd`, `cqrs` | Minimal API, EF Core, custom dispatcher | `postgres` |
| `dotnet` | `clean`, `ddd`, `cqrs`, `event-sourcing` | Marten on PostgreSQL, projections, optimistic concurrency | `postgres` |
| `dotnet` | `modular-monolith`, `ddd` | Catalog and Ordering modules, public contracts, integration events, one schema per module | `postgres` |
| `dotnet` | `vertical-slice` | Minimal API, one folder per feature, EF Core | `postgres` |
| `go` | `hexagonal`, `ddd` | `net/http` from the standard library, pgx | `postgres` |
| `go` | `microservices`, `event-driven` | Orders and billing with outbox, inbox and RabbitMQ (or HTTP without it) | `postgres`, `rabbitmq` |
| `kotlin` | `hexagonal`, `ddd` | Ktor, Gradle with wrapper, HikariCP, Konsist | `postgres` |
| `node` | `clean`, `ddd`, `cqrs` | NestJS with @nestjs/cqrs, TypeScript, pg, Vitest | `postgres` |
| `node` | `bff` | Orders API and BFF in Fastify, web in React; three processes | `postgres` |
| `node` | `layered` | Fastify, routes, services and repositories, pg | `postgres` |
| `node` | `trpc` | Monorepo with Next.js, tRPC, zod and pg | `postgres` |
| `java` | `hexagonal`, `ddd` | Spring Boot, Maven, JdbcClient, ArchUnit | `postgres` |
| `python` | `layered` | Django, Django REST Framework, uv | `postgres` |
| `react` | `feature-sliced` | Front end for the orders API: Vite, React Router, TanStack Query, FSD layer lint | — |
| `rust` | `hexagonal`, `ddd` | axum, tokio, sqlx | `postgres` |
| `php` | `hexagonal`, `ddd` | Symfony 7.4, Doctrine DBAL, PHPUnit | `postgres` |
| `python` | `hexagonal`, `ddd` | FastAPI, uv, psycopg | `postgres` |

`romanctl catalog` lists the templates and the known architectures: `bff`, `clean`, `cqrs`, `ddd`, `event-driven`,
`event-sourcing`, `feature-sliced`, `hexagonal`, `layered`, `microservices`, `modular-monolith`, `trpc` and
`vertical-slice`. Aliases such as `ports-and-adapters`, `vsa`, `modulith` and `fsd` also work.

After generating, `new` checks the tools the project needs and, in a terminal, offers to install them. With everything
installed, it runs the template's setup, such as `dotnet restore` or `go mod tidy`. With `--offline`, setup is left for
later. A setup failure does not delete the project.

### Where they live

Templates are not part of the binary. They live in this repository's [`templates/`](templates) directory, one
subdirectory per language, and are downloaded to `~/.local/share/romanctl/templates`, at the tag of the installed
version:

```
templates/
  go/
    projects/hexagonal-ddd/          a project that builds and tests in its own directory
    projects/microservices/
    generators/hexagonal-ddd/resource/
    generators/simple/handler/       generators for simple Go projects
  node/
    projects/clean-ddd-cqrs/  projects/layered/  projects/trpc/  projects/bff/
    generators/express/  generators/next/  ...
  react/  dotnet/  java/  kotlin/  php/  python/  rust/
```

```bash
romanctl templates                 # where they are and which version they came from
romanctl templates update          # downloads again, from the installed version's tag
romanctl templates update --ref refs/heads/main
```

The installer, `self-update` and `offline prep` download the templates by themselves, and so does the first `new` if
there are none yet. To develop a template, point `ROMANCTL_TEMPLATES` at the repository directory.

### How a template is built

Each template is a real project that builds and passes its tests inside its own directory, with the sample name
`sample-app`. Generating means copying the directory and replacing that name, in every spelling, with the project's:

| In the template | `my-shop` becomes |
|---|---|
| `SampleApp` | `MyShop` |
| `sampleApp` | `myShop` |
| `sample_app` | `my_shop` |
| `sample-app` | `my-shop` |
| `SAMPLE_APP` | `MY_SHOP` |
| `Sample App` | `My Shop` |
| `sample app` | `my shop` |
| `sampleapp` | `myshop`, as in `com.example.sampleapp` |

The replacement applies to contents and to paths, such as `src/SampleApp.Api`. The template directory is the variant
with every service. What changes when a service is left out goes in `_variants/`, which neither Go nor .NET compile:

```yaml
id: hexagonal-ddd
stack: go
arch: [ddd, hexagonal]
description: Go API with hexagonal architecture and DDD, using the standard library's HTTP
services: [postgres]                  # what --with accepts
requires: [go]                        # tools from romanctl tools
setup: [go mod tidy]                  # runs in the project directory after generating; a step may have when
run: go run ./cmd/api
test: go test ./...
url: http://localhost:8090
env:                                  # extra variables up writes to .env, per service
  postgres:
    PGHOST: "{{.Host}}"
publish:                              # CI and Dockerfile for romanctl publish
  kind: go
  vars: {Main: ./cmd/api}
conditional:                          # directories that only exist with or without a service
  - path: internal/adapters/postgres
    when: postgres
overlays:                             # files that replace the root ones when the service is left out
  - from: _variants/memory
    when: "!postgres"
```

A project with several programs, such as a monorepo with an API and a front end, declares `processes` instead of `run`:

```yaml
processes:
  - name: api
    dir: apps/api
    run: npm run dev
    url: http://localhost:3000
  - name: web
    dir: apps/web
    run: npm run dev
    url: http://localhost:5173
```

`run` and `up` start them all together, with the process name at the start of each log line, and `up -d` waits for each
`url` to respond. If a process exits, the others are stopped and the command fails right away, with the tail of the
log. `down` stops them all. Each process may have its own variables in `env:`, on top of the `.env` ones, such as a
different `PORT` for each.

With `processes`, the top-level `run` is optional.

A `setup` step may apply to a single variant. That way the variant without a service doesn't need to carry another
lockfile: it installs, and the package manager drops from the lock whatever its manifest doesn't ask for.

```yaml
setup:
  - run: npm ci
    when: postgres
  - run: npm install
    when: "!postgres"
```

With `uv`, `cargo` and `go mod tidy`, the lock already adjusts itself to the manifest; with Composer, the variant runs
`composer update` for the packages that were removed.

`run`, `test`, `publish.vars` and the processes' `dir` and `run` fields also go through the name replacement. The `env`
values are built by `up`, with the service address: `{{.Host}}`, `{{.Port}}`, `{{.User}}`, `{{.Password}}`,
`{{.Database}}` and `{{.URL}}`. On top of those, every project gets each service's standard variables, such as
`DATABASE_URL`.

A directory at `~/.config/romanctl/templates/<stack>/<id>/` takes priority over romanctl's templates. Each generated
project gets a `romanctl.yaml` with what `run`, `up`, `add` and `publish` use.

romanctl's tests generate each template with every combination of services and fail if a sample name is left over. CI
runs each template's tests in its own directory and also the generated project's, against a real PostgreSQL.

## Adding to a project

`romanctl add` generates code according to the project type, and `romanctl add --list` shows the options:

```bash
romanctl add resource Product                      # full CRUD
romanctl add domain Customer
romanctl add controller Customer
romanctl add command ArchiveProduct --in Product
romanctl add query ListActiveProducts --in Product
romanctl add event ProductArchived --in Product
romanctl add value-object Sku --in Product
romanctl add resource Person --plural People       # when the English plural rule gets it wrong
romanctl add resource Product --dry-run            # only shows
```

| Project | Generators |
|---|---|
| .NET Clean, DDD and CQRS | `resource`, `domain`, `application`, `repository`, `controller`, `command`, `query`, `event`, `value-object`, `test` |
| .NET simple | `controller`, `service`, `model` |
| Express | `route`, `middleware`, `service` |
| Next.js | `page`, `api`, `component` |
| React | `component`, `hook`, `page` |
| React feature-sliced | `entity`, `feature`, `page` |
| .NET event sourcing | `resource`, `domain`, `application`, `controller`, `test`, `event` |
| .NET modular monolith | `module`, `resource` (with `--in <module>`) |
| .NET Vertical Slice | `resource` |
| Go hexagonal and DDD | `resource` |
| Go microservices | `resource` (in the orders service) |
| Java hexagonal and DDD | `resource` |
| Kotlin hexagonal and DDD | `resource` |
| Node Clean, DDD and CQRS | `resource` |
| Node BFF | `resource` |
| Node layered | `resource` |
| Node tRPC | `resource` |
| PHP hexagonal and DDD | `resource` |
| Python hexagonal and DDD | `resource` |
| Python layered (Django) | `resource` |
| Rust hexagonal and DDD | `resource` |
| Go simple | `handler`, `service`, `model` |

Without `--force`, nothing is overwritten. Whatever needs registering in another file is inserted before a marker, such
as `// romanctl:endpoints`; without the marker, the command shows what to do by hand. Inserted lines take the marker's
indentation and keep their relative indentation, so nested blocks, such as YAML, go in whole. A block already in the
file is not inserted again.

Each generator is a directory at `templates/<language>/generators/<target>/<kind>`, with a `generator.yaml` and the
files in `files/`. The target is the template id (`go/hexagonal-ddd`) or the simple project type (`node/express`,
`node/next`, `react/vite`, `go/simple`, `dotnet/simple`). They use the sample entity `SampleItem` (and `SampleItems`,
`sample-item`, `sample_items`, `sample item`...), the `--in` aggregate `SampleParent` and the project `SampleApp`,
replaced as in the templates. A patch with `when: postgres` only applies with that service, and `after:` runs commands
after writing, such as `gofmt -w`. A directory at `~/.config/romanctl/generators/<language>/<target>/<kind>` takes
priority.

## Bringing up the environment

```bash
romanctl up                  # services, database, .env and the app, in the foreground
romanctl up shop -d          # in the background; waits for the app to respond
romanctl up --no-run         # only prepares the environment
romanctl logs shop
romanctl down shop           # or --all, or --keep-services
```

`up` reads the services from `romanctl.yaml`, creates the project's database and appends to `.env` only what is
missing, without changing what you edited. `.env` goes into `.gitignore`, and the variables apply to the app, except
those already set in the environment. `romanctl run` also loads `.env`, just without starting the services. If the port
is taken, the command stops and says by which project, when that's the case.

`down` stops only the services that `up` started and that no other running project uses. Whatever was already running
before stays running. Projects without `romanctl.yaml` have no services: `up` just runs the app.

## Publishing

```bash
romanctl publish                                   # private repository
romanctl publish shop --public --description "Shop"
romanctl publish --dry-run
```

Generates CI and a Dockerfile for the project, without overwriting existing ones, commits, creates the repository with
`gh` and pushes. If the repository already exists, it only pushes the changes. Files such as `.env` and private keys
cancel the publish, as do secrets found by `gitleaks`, if it is installed.

## Services

Outside proot, services run in Docker or Podman, with ports bound only to `127.0.0.1` and data in named volumes.
code-server runs as a local process.

`romanctl status` and the dashboard show only what is installed: Docker, when present, and the services whose
`romanctl-<service>` container exists. Having Docker doesn't install anything, nor does an image pulled by another
project: `romanctl install <service>` pulls the image, if missing, and creates the stopped container. The rest shows up
on a single line, or with `romanctl status --all`.

Docker shows up as a service because the others depend on it. Starting and stopping the daemon uses systemd (or
`service`) with `sudo`; stopping asks for confirmation, because it brings down every container on the machine, and
`stop all` never includes it. With Docker Desktop, the daemon is started and stopped by Docker Desktop itself.

| Service | Address |
|---|---|
| `code` | `http://127.0.0.1:8080` |
| `postgres` | `postgres://postgres:postgres@127.0.0.1:5432/postgres` |
| `mongo` | `mongodb://127.0.0.1:27017` |
| `redis` | `redis://127.0.0.1:6379` |
| `mysql` | `mysql://root:mysql@127.0.0.1:3306` |
| `mariadb` | `mysql://root:mariadb@127.0.0.1:3307` |
| `sqlserver` | `Server=127.0.0.1,1433`; `amd64` only; the image accepts the SQL Server Developer license |
| `rabbitmq` | `amqp://guest:guest@127.0.0.1:5672`; management UI at `http://127.0.0.1:15672` |
| `kafka` | `127.0.0.1:9092` |

The passwords are fixed and meant only for local development.

## System

```bash
romanctl system           # CPU, memory, disk, battery and processes
romanctl system battery   # level, watts, temperature and estimated time left
romanctl system watch     # measures for a while and estimates what your processes draw
```

Time left comes from the remaining charge divided by the current, and each reading goes into a history that gives the
real drop in %/h. On Linux the battery comes from `/sys/class/power_supply`, on macOS from `pmset` and `ioreg`, and on
Termux from `termux-battery-status`, which requires the Termux:API app.

None of these systems tell an ordinary program the power draw per process. `watch` uses regression to work out how much
the draw rises with CPU usage and refuses to give a number when the data doesn't support one.

## After a crash

Each `start` records the service in the state file under `~/.local/share/romanctl` and each `stop` removes it. After a
crash or a reboot, `romanctl restore` brings back only what was running. Projects in `romanctl run` don't come back by
themselves, because they need their own terminal: `restore` shows the command for each one.

## Without internet

`romanctl offline prep` downloads, while online, React and Next templates with `node_modules`, npm, Go and NuGet caches,
and code-server extensions. It also configures `GOPROXY` to fall back to the local cache when the network fails.

Without internet, `romanctl new` uses the templates and caches on its own. To add packages:

```bash
npm install <package> --offline
go get <module>
dotnet add package <package>
```

`romanctl offline test` creates and builds each project type with the network blocked by a dead proxy. This test is not
a substitute for airplane mode.

## Automation

`--json` works on `status`, `list`, `catalog`, `new`, `add`, `publish`, `tools`, `system`, `platform`, `claude` and
`config`, and progress messages go to stderr. Exit codes: 0 success, 1 execution failure, 2 incorrect usage. Without a
terminal, commands never open a menu: pass everything as arguments.

## Claude Code

```bash
romanctl claude install      # writes the skills to ~/.claude/skills
romanctl claude status       # installed, outdated, missing or foreign
romanctl claude uninstall
```

The skills ship inside the binary, in `internal/claude/skills`, and teach Claude Code to do everything through
romanctl: a general one, with the rules for use without a terminal, and one per area (tools, projects, environment,
generators, services, system, publishing and maintenance). romanctl marks the directories it creates with a `.romanctl`
file and only replaces or deletes those; a hand-made skill with the same name is only replaced with `--force`. After a
`self-update`, run `romanctl claude install` again.

## Termux and proot

`romanctl` detects `proot-distro` and changes behavior only in that environment:

- **Containers don't run**, because Android doesn't allow namespaces and cgroups. `docker`, `podman`, `kind` and the
  like are refused, with the reason. `docker-cli` works against a remote Docker.
- **Postgres, Mongo and Redis run as local processes.** The other services are unavailable.
- **PostgreSQL** needs the `pgshim.c.txt` shim, loaded through `LD_PRELOAD`. It fixes the owner `stat()` reports and
  emulates SysV shared memory. It is a workaround: back up whatever matters.
- **MongoDB** stays on the 7.0 series, because the newer versions tested fail to allocate memory on the Android kernel.
- **.NET** gets a wrapper with a heap limit and the host path for `dotnet test`.
- **Services stop** when the Ubuntu session that started them ends (`--kill-on-exit`).

## Code layout

| Directory | Responsibility |
|---|---|
| `cmd/romanctl` | Entry point |
| `internal/cli` | Commands and menus |
| `internal/tools` | Tool registry and installation, plus each tool's version and README |
| `internal/tui` | The dashboard |
| `internal/claude` | The Claude Code skills and their installation |
| `internal/services` | code-server, PostgreSQL, MongoDB, Redis and `restore` |
| `internal/projects` | Create, list, run and remove simple projects |
| `internal/scaffold` | Architecture templates and generators |
| `internal/publish` | CI, Dockerfile and publishing to GitHub |
| `internal/offline` | Preparation and testing without internet |
| `internal/manifest` | Each project's `romanctl.yaml` |
| `internal/state` | What should be running, for `restore` |
| `internal/platform` | Detects system, architecture, package manager, root, WSL and proot |
| `internal/browser` | Opening URLs on each system |
| `internal/i18n` | Language detection and the Portuguese translation |
| `internal/webui` | The embedded web interface server |
| `internal/sys`, `internal/ui`, `internal/config` | Processes and network, output and prompts, paths |

```bash
go test ./...
```

CI runs on Linux and macOS, on both architectures, and in Debian, Fedora, Arch, openSUSE and Alpine containers.
`scripts/toolsweep.py` installs and removes every tool; in CI, run it by hand with `sweep`. A `v*` tag publishes the
binaries and checksums to a release.
