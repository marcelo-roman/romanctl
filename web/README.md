# romanctl web-ui — front-end

React + Vite + TypeScript SPA served by `romanctl web-ui`. The build goes to
`../internal/webui/dist`, which the Go binary embeds with `go:embed`.

## Requirements

- Node.js 22.12+ and npm (versions are pinned in `package.json`/`package-lock.json`)
- Go, to build and run the backend

## Install

```bash
cd web
npm ci
```

## Develop

The Vite dev server proxies `/api` to a running `romanctl web-ui` and injects
the backend token into the page, so the API calls are authorized.

1. Start the backend (any free port; temporary folders keep the real machine untouched):

   ```bash
   go build -o /tmp/rctl-web/romanctl ./cmd/romanctl
   ROMANCTL_PROJETOS=/tmp/rctl-web/proj ROMANCTL_TEMPLATES=$PWD/templates \
     /tmp/rctl-web/romanctl web-ui --port 7790
   ```

2. Start Vite pointing at it:

   ```bash
   cd web
   ROMANCTL_WEBUI_URL=http://localhost:7790 npm run dev
   ```

   On each page load the dev server fetches `/` from the backend, reads the
   `<meta name="romanctl-token">` it injects and adds the same meta to the dev
   page. The proxy drops the `Origin` header and rewrites `Host`, because the
   backend only accepts its own origin.

   If the backend is not reachable when the page loads, restart it and reload.
   To pin a token instead, set `VITE_ROMANCTL_TOKEN=<token>` (the value of the
   meta tag in the backend page).

Variables:

| Variable              | Default                  | Use                                    |
| --------------------- | ------------------------ | -------------------------------------- |
| `ROMANCTL_WEBUI_URL`  | `http://localhost:7777`  | backend the dev proxy talks to         |
| `VITE_ROMANCTL_TOKEN` | read from the backend    | fixed token for the dev page           |

## Test

```bash
npm test          # tsc --noEmit, eslint and vitest (jsdom + MSW)
npm run test:watch
```

## Build

```bash
npm run build     # writes ../internal/webui/dist (emptied first)
cd .. && go build ./... && go test ./internal/webui/
```

`index.html` must keep its `</head>`: the server injects the token meta right
before it. Assets are hashed under `assets/` and served with a long cache.
The page runs under a strict CSP (`default-src 'self'`), so everything is
bundled — no CDN fonts, scripts or styles.

## Layout

```
src/
  api/         typed client (token header, backend error), types, TanStack Query hooks
  i18n/        pt/en dictionary, provider and useI18n
  lib/         pure logic: formatting, validation, versions, filters, job bodies
  jobs/        job tracking (polling, toast and query invalidation when a job ends)
  feedback/    toasts
  components/  dialog with focus trap, confirm, log viewer, badges, fields
  layout/      sidebar, header, shell
  features/    one folder per screen: dashboard, projects, services, tools, templates, jobs
  test/        MSW server, fixtures, render helper
```
