---
name: romanctl-maintenance
description: Maintains romanctl itself — update to the new version (self-update), change the language, prepare the machine to work without internet (offline prep/test) and install or update these Claude skills. Use when the user asks to update romanctl, change the language, work offline, or when the skills are outdated.
---

# romanctl maintenance

Follow the general rules of the `romanctl` skill.

## Version and updates

```bash
romanctl version
romanctl self-update --check --json        # current, latest, updateAvailable
romanctl self-update --yes --json          # updates; verifies the SHA-256 and tests before swapping
romanctl self-update --version v0.1.0 --yes --json   # goes back to a specific version
```

After updating, run `romanctl claude install` so the skills follow the new version.

## These skills

```bash
romanctl claude status --json      # installed, outdated, missing or foreign, per skill
romanctl claude install --json     # installs or updates
romanctl claude uninstall --json   # removes only the ones romanctl installed
```

romanctl only touches folders it created itself (they have the `.romanctl` file). A folder with the
same name created by hand shows up as `foreign` and is only overwritten with `--force`; ask the
user first.

## Language

```bash
romanctl config lang pt    # or en, or auto to follow the system
```

For a single call: `--lang en`. Messages change; JSON keys do not.

## Without internet

```bash
romanctl offline status
romanctl offline prep [--force]   # downloads templates, caches and extensions; needs internet
romanctl templates update --json  # downloads the project templates and generators again
romanctl offline test             # creates and builds each project type with the network blocked
```

`offline prep` and `offline test` take a while; warn the user and run them in the background. Once the preparation
is done, `romanctl new <type> <name> --offline` creates projects without network.
