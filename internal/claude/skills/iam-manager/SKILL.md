---
name: iam-manager
description: Administers Keycloak through the IAM Manager REST API with a client_credentials credential — create, search, block and delete users, generate invite and password reset links, register applications and grant access to them, register operators, create and rotate API keys, read login events and the audit log, apply themes. Use when the request involves users, passwords, applications, permissions, API keys or login for an app that uses Keycloak, or when IAM_CLIENT_ID and IAM_CLIENT_SECRET show up.
---

# IAM Manager

IAM Manager is the administration gateway to Keycloak. Everything goes through its API, never through the Keycloak
admin console or the Admin API directly: it validates permissions, protects the `master` realm and
records every change in the audit log with your `client_id`.

There are two realms: `private` (operators and internal tools, login only through criarch's Google
Workspace) and `public` (application users). Realms cannot be created or deleted through the API.

The source of truth is the specification: `GET $IAM_API_URL/api/openapi.json`. The full guide, with
recipes, is at `GET $IAM_API_URL/api/guide.md` and in [references/api.md](references/api.md).

## Credential

The credential is an **API client** (client_credentials) and comes from environment variables:

| variable | example |
| --- | --- |
| `IAM_API_URL` | `https://iam-api.marceloroman.com.br` |
| `IAM_TOKEN_URL` | `https://auth.marceloroman.com.br/realms/private/protocol/openid-connect/token` |
| `IAM_CLIENT_ID` | `iam-api-...` |
| `IAM_CLIENT_SECRET` | secret |

If they are not in the environment, look for `~/.config/iam-manager/credentials.env` and load it with
`set -a; . ~/.config/iam-manager/credentials.env; set +a`. If there is no credential, ask the
user; do not make one up or look elsewhere.

With `romanctl` installed, `romanctl iam` makes the same calls with that credential and handles the
token: `romanctl iam credentials --json` tells whether it exists, and `romanctl help iam` lists the commands.
To save it, the user runs `! romanctl iam credentials set --client-id <id>`, which asks for the secret
without echo.

**Never** show, log, commit or repeat the `IAM_CLIENT_SECRET`, the token, temporary passwords or
invite links in a response, log or file. To store the credential, use the file above with
`chmod 600`.

## Token

```bash
iam_token() {
  curl -sf -X POST "$IAM_TOKEN_URL" -d grant_type=client_credentials \
    -d client_id="$IAM_CLIENT_ID" -d client_secret="$IAM_CLIENT_SECRET" | jq -r .access_token
}
TOKEN=$(iam_token)
iam() { curl -s -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' "$@"; }
```

The token is valid for 5 minutes. On 401, get a new one **once**; if it persists, stop and report.

## Before anything else

```bash
iam "$IAM_API_URL/api/me"
```

`globalPermissions` applies to all realms; `realmPermissions` lists the delegated realms
(`public:admin` or `public:viewer`). Do not
attempt what the permissions do not cover: a 403 is a final answer, not something to work around.

## Rules

- **Deletion cannot be undone.** Before deleting a user, API key, API client, web app or theme
  version, confirm with the user, stating exactly what will be gone.
- **New accounts with `credential: "link"`.** The link goes to the person, it does not stay with you. Temporary
  password only if the user asks.
- **Users need** first name, last name and email in their profile, otherwise login gets stuck at
  "Account is not fully set up".
- **Joining the realm grants access to no application.** Access is the user's roles in each
  application (`PUT .../users/{id}/access/{app}`); an empty list revokes.
- **A new operator is registered before their first login** (`POST /api/operators` with the
  Workspace email); `private` refuses anyone who was not registered.
- **PUT of roles/groups (`/membership`) and of permissions takes the complete desired state.** Read the
  current one, change it and send everything; do not send only what changes.
- **Unknown fields are rejected (422).** Send only what the specification defines.
- **User lists are paginated** (`first`, `max` up to 100; the response has `total`).
- **Secrets appear once** (API key, rotation, temporary password, link). Hand them to the user and do not
  keep or repeat them afterwards.
- The `master` realm is never accessible. `private` holds the operators and only accepts `iam-admin`; only
  touch it if the user explicitly asks.

## Quick map

| to | call |
| --- | --- |
| realms and settings (read-only) | `GET /api/realms`, `GET /api/realms/{realm}/settings` |
| users | `GET /api/realms/{realm}/users?search=&first=&max=`, `POST`, `GET/PUT/DELETE .../users/{id}` |
| user roles and groups | `GET /api/realms/{realm}/catalog`, `PUT .../users/{id}/membership` |
| password reset | `POST .../users/{id}/credential` with `{"mode":"link"}` |
| applications | `GET` / `POST /api/realms/{realm}/applications`, `POST .../{app}/roles`, `PUT .../{app}/theme` |
| user access to applications | `GET .../users/{id}/access`, `PUT .../users/{id}/access/{app}` with `{"roles":[...]}` |
| web apps (clients with login) | `GET /api/realms/{realm}/web-apps`, `PUT/DELETE .../web-apps/{clientId}` |
| operators | `GET /api/operators?q=`, `POST /api/operators`, `PUT /api/operators/{id}/roles` |
| API clients (this kind of credential) | `GET/POST /api/api-clients`, `PUT/DELETE .../{id}`, `POST .../{id}/rotate` |
| an app's API keys | `GET/POST /api/realms/{realm}/api-keys`, `POST .../{id}/rotate` |
| login events | `GET /api/realms/{realm}/events`; enable: `PUT .../events/config` |
| audit log | `GET /api/audit?limit=` |
| themes | `POST /api/themes` (multipart), `PUT /api/realms/{realm}/themes` |

Complete body examples for each case in [references/api.md](references/api.md).

## Errors

Always `{"error":"..."}`: 401 token, 403 no permission, 404 does not exist, 409 conflict (name already
exists, resource in use), 422 validation (read the message), 429 wait a minute, 503 Keycloak down —
retry with a delay.
