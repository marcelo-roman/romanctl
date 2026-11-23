---
name: romanctl-system
description: Monitors the machine through romanctl — CPU, memory, swap, disk, battery (level, watts, remaining time), top-consuming processes, estimated power draw in watts, operating system, architecture and detected capabilities (WSL, containers, virtualization). Use when the user asks whether the machine is under load, what is consuming resources, how much battery is left, or before installing something large.
---

# Monitoring through romanctl

Follow the general rules of the `romanctl` skill. Always read with `--json` and summarize for the
user in sentences; do not dump the JSON.

## Snapshot now

```bash
romanctl system status --json
```

Returns `platform`, `cores`, `systemCpuPercent` (null when the machine does not report it; then use
`visibleProcessesCores`), `memory` (`totalBytes`, `availableBytes`, `swapTotalBytes`,
`swapFreeBytes`), `disk` (`path`, `sizeBytes`, `freeBytes`), `battery` (when present) and
`top` (processes with `pid`, `name`, `cpuPercentOfOneCore`, `rssBytes`).

## Battery

```bash
romanctl system battery --json
```

Level, power draw in watts, temperature and remaining-time estimate, when the system reports them.

## Measure over a period

```bash
romanctl system watch --seconds 60 --interval 5 --json
```

**Always with `--seconds`**: without it, the command asks for the duration. It measures power and CPU together
and estimates how many watts the user's processes use (`fit.estimatedW`, `fit.meanPowerW`). For long measurements (minutes), run
it in the background and tell the user how long it will take.

## The machine

```bash
romanctl platform --json   # os, arch, wsl, proot, root, manager, binDir, containers, virtualization, gui
romanctl doctor            # versions of the main tools
```

Use `platform` before suggesting anything that depends on the system: `containers: false` means
Docker services will not start; `manager` tells which package manager exists.

## Interpreting

- Memory: what matters is `availableBytes`.
- `cpuPercentOfOneCore` goes above 100% for multi-threaded processes; divide by `cores`
  for the machine's share.
- Less than ~10% free disk gets in the way of Docker images and builds; warn about it.
- A Claude or code-server process at the top is normal; name it before
  suggesting ending anything, and do not kill processes unless asked.
