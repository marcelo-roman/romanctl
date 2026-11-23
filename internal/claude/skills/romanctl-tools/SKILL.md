---
name: romanctl-tools
description: Installs, removes, updates and inspects development tools through romanctl (kubectl, helm, terraform, node, go, dotnet, gh, docker CLI, database clients, linters, cloud CLIs, about 170). Use when the user asks to install a program, when a command fails with "command not found", or to find out the installed version, available version, source and README of a tool.
---

# Tools through romanctl

Follow the general rules of the `romanctl` skill (no terminal, `--json`, `--yes`).

## Discover

```bash
romanctl tools list --plain --json                 # what runs on this machine, with state and version
romanctl tools list --all --filter <text> --json   # includes what does not run here, with the reason
romanctl tools list --category k8s,db --json       # by category
romanctl tools categories --json                   # categories and counts
romanctl tools info <name> --json                  # installed, available, source, page
romanctl tools info <name>                         # the same, with the README in markdown
romanctl tools versions <name> --json              # installable versions, newest to oldest
```

Categories: `sdk`, `node-tools`, `python-tools`, `build`, `git`, `cloud`, `k8s`, `iac`,
`containers`, `security`, `db`, `observability`, `network`, `shell`, `ai`. Portuguese aliases
also work (`banco`, `nuvem`, `linguagens`...).

`tools list --json` includes `latest` when the latest version was checked in the last 6 hours.
`tools info` answers `upToDate` and `upstream.latest`: the version an installation would bring
right now, checked against the same source the installer uses (GitHub releases, apt/dnf/pacman,
Homebrew, npm, PyPI or the project's official installer).

## Install

```bash
romanctl tools install <name> [<name>...] --yes --json
romanctl tools install <group> --yes --json     # the whole group
romanctl tools install <name> --force --yes     # reinstalls: this is how you update
romanctl tools install <name> --version <version> --yes --json   # a specific version, even an older one
romanctl tools install <names> --dry-run        # the plan, without installing
```

### Specific version

When the user asks for a version ("terraform 1.9", "node 20"), check
`tools versions <name> --json` and pick the matching `id` (or `label`); for a partial version,
prefer the newest of that line and say which one you chose. `--version` applies to one tool
at a time, and changing the version of something installed is already a reinstall, without `--force`.

Not every tool allows choosing: with Homebrew on macOS and with some custom
installers (aws, gcloud, dart, swift, rust, docker-cli), `versions` answers that only the
latest exists. Tell the user that instead of looking for another way.

- Without `--yes`, a group would open a picker; without a terminal, `--yes` is required
  for groups.
- The JSON result has `results[]` with `action` (`installed`, `skipped`, `failed`,
  `unavailable`) and `version`. `skipped` means it was already installed.
- After installing, if romanctl warns to add a folder to PATH, pass the warning on to the
  user; do not edit `.zshrc`/`.bashrc` without asking.
- Installation downloads from the internet and may ask for `sudo` (system packages). If it fails due to
  permissions, tell the user and suggest `! romanctl tools install <name>` for them to run.

## Remove

```bash
romanctl tools uninstall <name> --yes --json
romanctl tools uninstall <names> --dry-run
```

`--yes` refuses `all` on purpose. Some tools warn about what stops working
(`tools info` shows it under "When removing"); confirm with the user before removing SDKs.

## When the tool is not in the catalog

An empty `romanctl tools list --all --filter <name> --json` means romanctl does not
know it. Tell the user and offer to install it outside romanctl this time or to add it to
romanctl's catalog (`internal/tools/catalog_*.go` in the romanctl repository).
