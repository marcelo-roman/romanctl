---
name: romanctl
description: Rules for operating this machine through romanctl, the CLI that centralizes tool installation, project creation and execution (web, .NET, Go, Node, Python, Java, Kotlin, PHP, Rust), Docker services (Postgres, Redis, Kafka...), code generation, GitHub publishing and machine monitoring. Use before installing any program, creating or running a project, starting a database, checking CPU/memory/battery or publishing a repository.
---

# romanctl

On this machine, romanctl is the default path to the development environment. Before
reaching for `apt`, `brew`, `npm -g`, `docker run`, `dotnet new`, `npx create-*` or one-off
scripts, check whether romanctl does it. It records what it installed, what it started and what
each project uses, and that is what makes it possible to tear down and replicate the environment later.

## How to call it

Claude runs without an interactive terminal. So:

- **Pass everything as arguments.** Without a terminal, romanctl asks nothing; if something is
  missing, it fails with `usage: true` and says what is missing.
- **Use `--json` to read results.** Errors also come as JSON: `{"ok": false, "error": "...", "usage": bool}`.
  Exit code 2 is a usage error (fix the arguments); 1 is an operation failure.
- **Use `--yes` when the command asks for confirmation** and you already have the user's approval.
  `--dry-run` shows the plan without executing; use it before anything destructive or large.
- **romanctl only runs with a login.** If a command fails with "only works with a login" or "session
  expired", do not try to work around it: login opens the browser, so ask the user to run
  `! romanctl login` (signs in with Google, using a criarch account). `romanctl whoami --json` tells whether
  there is a session. To administer Keycloak (users, applications, operators, API keys), `romanctl iam`
  uses an IAM Manager API client, not the login: see the `iam-manager` skill and `romanctl help iam`.
- **When in doubt, `romanctl help <command> [subcommand]`.** The help is the source of truth for
  options; this skill is a map.

### Never run these, because they block the process or open a full screen

| Do not run                              | Instead                                                  |
| --------------------------------------- | -------------------------------------------------------- |
| `romanctl` with no arguments, `ui`, `menu` | the specific subcommand                               |
| `romanctl run <project>`                | `romanctl up <project> -d --json`                        |
| `romanctl up <project>` without `-d`    | `romanctl up <project> -d --json`                        |
| `romanctl logs <target>` (follows the log) | `timeout 15 romanctl logs <target>` or read the log file |
| `romanctl system watch` without `--seconds` | `romanctl system watch --seconds 60 --json`          |
| `romanctl code`, `romanctl open`        | only when the user asks to open something in the browser |
| `romanctl web-ui`                       | only when the user asks for the web interface; it blocks the terminal, so run it in the background |

## Skill map

| To                                                        | Skill                  |
| --------------------------------------------------------- | ---------------------- |
| install, remove, update or inspect tools                  | `romanctl-tools`       |
| create, list or remove projects; templates                | `romanctl-projects`    |
| start and stop a project's environment; app logs          | `romanctl-dev`         |
| generate code inside a project (resource, page...)        | `romanctl-add`         |
| Docker, databases, queues and code-server                 | `romanctl-services`    |
| CPU, memory, disk, battery, processes, platform           | `romanctl-system`      |
| publish to GitHub with CI and Dockerfile                  | `romanctl-publish`     |
| update romanctl, language, offline mode, these skills     | `romanctl-maintenance` |
| Keycloak users, applications, passwords and API keys      | `iam-manager`          |

## When romanctl does not cover the request

Today it has no templates for mobile (Expo, Flutter, React Native) or desktop, nor for languages
other than .NET, Go, Node, React, Python, Java, Kotlin, PHP and Rust. If the request falls outside
what `romanctl catalog` and `romanctl types` show:

1. Tell the user romanctl does not do that yet.
2. Offer to do it manually this time **or** to add support for it to romanctl itself.
3. Do not install or create anything outside it without saying so, because the goal is for everything to go through it.

The same applies to tools: `romanctl tools list --all --filter <name> --json` before
installing by any other means.

## Where things live

- Projects: in the folder shown at the end of `romanctl help` (default `~/projetos`).
- Tool binaries: `binDir` in `romanctl platform --json`.
- State of what is running: romanctl keeps it; query it with `romanctl status --json` and
  `romanctl list --json`, without reading internal files.
