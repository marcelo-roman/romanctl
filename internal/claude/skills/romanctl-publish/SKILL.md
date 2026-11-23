---
name: romanctl-publish
description: Publishes a project to GitHub through romanctl — generates CI (GitHub Actions) and a Dockerfile according to the template, refuses sensitive files such as .env, commits and creates the repository (private by default) with gh. Use when the user asks to publish, push to GitHub or create a project's repository.
---

# Publishing through romanctl

Follow the general rules of the `romanctl` skill. Publishing is an external action: **only on the user's
explicit request**, and show the plan first.

## 1. Plan

```bash
romanctl publish <project> --dry-run --json
```

Shows `name`, `public`, `remoteExists`, `remote` and the files that would be generated
(`generated`). Pass this on to the user: repository name, whether it is public or private, and what
CI and Dockerfile it includes.

## 2. Publish

```bash
romanctl publish <project> --yes --json
romanctl publish <project> --yes --description "Sample store" --json
romanctl publish <project> --yes --public --json   # only if the user asks for public
romanctl publish <project> --yes --no-ci --json    # no CI or Dockerfile
```

- Without a terminal, `--yes` is required; it also skips the visibility questions, so the
  repository is created **private** unless you pass `--public`.
- The result has the repository `url` and `committed`.
- If romanctl refuses because of a sensitive file (`.env`, keys), do not force it or
  delete the file: tell the user which one it is and suggest adding it to `.gitignore`.
- Requires an authenticated `gh`. If it is missing: `romanctl tools install gh --yes` and ask the
  user to run `! gh auth login`, which is interactive.
