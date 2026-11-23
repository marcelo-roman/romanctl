---
name: romanctl-services
description: Starts, stops, restarts, installs and monitors romanctl's local services — the Docker daemon, the databases and queues in containers (PostgreSQL, MongoDB, Redis, MySQL, MariaDB, SQL Server, RabbitMQ, Kafka) and code-server. Use when the user needs a local database or queue, a connection string, wants to know what is running, or wants to start or stop Docker.
---

# Services through romanctl

Follow the general rules of the `romanctl` skill. Use these commands instead of `docker run` or
`docker compose` for development databases and queues.

Keys: `docker`, `code`, `postgres`, `mongo`, `redis`, `mysql`, `mariadb`, `sqlserver`,
`rabbitmq`, `kafka`.

## State and connection

```bash
romanctl status --json         # only what is installed on this machine
romanctl status --all --json   # includes what can be installed
```

Each item has `installed`, `running`, `url` (the connection string or address, with development
user and password), `wanted` (romanctl wants it running), `engine` (it is the
Docker daemon) and `unsupported`. Use the `url` to configure the project.

A container service only counts as installed when the `romanctl-<service>` container
exists. Having Docker, or the image pulled by another project, is not enough: if the service does not
show up in `status --json`, install it before using it.

## Start and stop

```bash
romanctl install <service>     # pulls the image, if missing, and creates the container stopped; once
romanctl start <service>
romanctl stop <service>
romanctl restart <service>
romanctl start all             # all installed ones
```

Without an argument, these commands would open a picker: always pass the service or `all`. `all`
never includes Docker.
In a project, prefer `romanctl up <project>` (`romanctl-dev` skill), which starts only what the
project uses and knows how to stop it afterwards.

`stop` on a service that a running project uses brings that project down underneath it; check
`romanctl list --json` (`up.services`) first and warn the user.

## Docker

The daemon shows up as the `docker` service. `running: false` on it explains why no
container starts; tell the user that before trying anything else.

```bash
romanctl start docker              # systemd or service, with sudo
romanctl stop docker --yes         # brings down ALL containers on the machine
```

- Starting and stopping asks for `sudo`. Without a terminal, it fails with "needs the sudo password": ask the
  user to run `! romanctl start docker`.
- **Stopping Docker brings down all containers on the machine**, including those of other projects
  that are not romanctl's. Only on explicit request, after warning; romanctl refuses without
  `--yes` outside a terminal.
- With Docker Desktop, the daemon is started and stopped by Docker Desktop.

## Logs

`romanctl logs <service>` follows the log and does not exit:

```bash
timeout 15 romanctl logs postgres
```

## After a crash or reboot

```bash
romanctl restore status     # what was saved as running and what went down
romanctl restore            # starts it again; without a terminal, restores directly
romanctl restore save       # records what is running now as the state
romanctl restore clear      # forgets the state
```
