---
name: romanctl-dev
description: Starts and stops a project's development environment through romanctl (Docker services, database, .env and the app in the background), and reads the app logs. Use to run, test in the browser, restart or stop a project, or when the user says "start it", "run it", "bring it down", "stop the project" or "what is running".
---

# Development cycle through romanctl

Follow the general rules of the `romanctl` skill. The project can be passed by name; without a name,
romanctl uses the project in the current folder.

## Start

```bash
romanctl up <project> -d --json        # services + database + .env + app in the background
romanctl up <project> --no-run --json  # only the environment (services, database, .env), without the app
```

- **Always `-d` or `--no-run`.** Without them, the command only returns when the app stops.
- `-d` waits for the app to respond. The JSON has `url`, `pid`, `log`, `services`,
  `startedNow` (what was started now), `envAdded` (variables added to `.env`) and
  `alreadyUp`.
- `.env` only gets what is missing; whatever the user changed stays. Do not overwrite `.env` by hand
  with romanctl values.
- A busy port is refused, naming who is using it: pass that on to the user instead of killing processes.
- Without a terminal, `up` does not install a service by itself: if it fails with "is not installed — run:
  romanctl install <service>", run that `install` and repeat the `up`.

## Check what is running

```bash
romanctl list --json      # projects; the running ones have up.alive, up.url and up.services
romanctl status --json    # services: installed, running, url
```

To test the app, use the `url` returned by `up` (for example with `curl`).

## Logs

`romanctl logs <project>` **follows the log and does not exit**. Prefer reading the file: the path
comes in the `log` field of `up -d --json`.

```bash
tail -n 100 <log>
timeout 15 romanctl logs <project>     # if you need to follow it for a while
```

## Stop

```bash
romanctl down <project> --json                  # app + services only it was using
romanctl down <project> --keep-services --json  # only the app
romanctl down --all --json                      # everything romanctl started
```

`down` does not stop a service that was already running before `up`, nor a service another
project uses: the JSON separates `servicesStopped` from `servicesKept`.

## Restart the app after changing code

```bash
romanctl down <project> --keep-services --json && romanctl up <project> -d --json
```
