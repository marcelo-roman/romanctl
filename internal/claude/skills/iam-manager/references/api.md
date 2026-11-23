# IAM Manager API — usage guide

REST API to administer the Keycloak realms `private` (operators and internal tools) and `public`
(application users): users, credentials, applications and access, API keys, social
login, web apps, events, themes, operators and audit. This guide is for people and for AI agents.

- Full specification: `GET {API}/api/openapi.json` (or `.yaml`)
- This guide: `GET {API}/api/guide.md`

## 1. Addresses

| name | production value |
| --- | --- |
| `API` | `https://iam-api.marceloroman.com.br` |
| `TOKEN_URL` | `https://auth.marceloroman.com.br/realms/private/protocol/openid-connect/token` |

## 2. Authentication (client_credentials)

You receive a `client_id` and a `client_secret` for an **API client** (created in
IAM Manager → Operators → API clients, in the `private` realm). Exchange them for an access token
with audience `iam-manager-api`:

```bash
TOKEN=$(curl -s -X POST "$TOKEN_URL" \
  -d grant_type=client_credentials \
  -d client_id="$IAM_CLIENT_ID" \
  -d client_secret="$IAM_CLIENT_SECRET" | jq -r .access_token)
```

- The token is valid for 5 minutes (`expires_in`). Reuse it until it is close to expiring and request another; there is no
  refresh token in client_credentials.
- Send it on every call: `Authorization: Bearer $TOKEN`.
- **Never** record the secret or the token in a log, commit, issue or chat response.

Check who you are and what you can do:

```bash
curl -s -H "Authorization: Bearer $TOKEN" "$API/api/me"
```

`globalPermissions` applies to all realms; `realmPermissions` lists delegated realms. The roles
are client roles of `iam-manager-api`: global `iam-admin`, `iam-user-admin`, `iam-app-admin`,
`iam-theme-admin` and `iam-auditor`, and the delegations `public:admin` and `public:viewer`.

## 3. Conventions

- JSON everywhere (`Content-Type: application/json`), except theme upload (`multipart/form-data`).
- Unknown fields in the body are rejected (422). Send only what the specification defines.
- User lists are paginated: `first` (offset) and `max` (up to 100); the response has `total`.
- The PUTs of roles/groups and of permissions take the **complete desired state**, not a diff.
- `realm` is `private` or `public`; realms cannot be created or deleted through the API.
- The `master` realm is never accessible. `private` (operators) only with `iam-admin`.
- Every change goes into the audit log (`GET /api/audit`) with the `client_id` as author.

### Errors

The body is always `{"error": "message"}`.

| status | meaning | what to do |
| --- | --- | --- |
| 401 | token missing, expired or invalid | get another token |
| 403 | missing permission for that realm/action | do not insist; report |
| 404 | realm, user or resource does not exist | check the identifier |
| 409 | conflict (name already exists, version in use) | choose another name or resolve the usage |
| 422 | validation (field, weak password, rule) | read the message and fix it |
| 429 | rate limit | wait a minute |
| 503 | Keycloak unavailable | retry with a delay |

## 4. Recipes

Variables used: `API`, `TOKEN`. Use `H='Authorization: Bearer '$TOKEN`.

### Realms

```bash
curl -s -H "$H" "$API/api/realms"                                       # list
curl -s -H "$H" "$API/api/realms/public/settings"                       # settings (read-only)
```

Settings come from Terraform; the API only shows them.

### Users

```bash
# create with an invite link (the person sets the password)
curl -s -X POST -H "$H" -H 'Content-Type: application/json' "$API/api/realms/public/users" -d '{
  "profile": {"username":"ana","email":"ana@acme.com","firstName":"Ana","lastName":"Silva","enabled":true,"emailVerified":true},
  "credential": "link", "roles": [], "groups": []
}'
```

`credential`: `link` (returns `credential.link`, single use, 72 h), `temporary` (returns
`credential.temporaryPassword`, mandatory change at first login), `email` (Keycloak sends it; requires
SMTP in the realm) or `none`. The profile needs first name, last name and email for login to work.

```bash
curl -s -H "$H" "$API/api/realms/public/users?search=ana&first=0&max=25"   # search
curl -s -H "$H" "$API/api/realms/public/users/$ID"                         # detail
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"roles":["billing-reader"],"groups":["<group-id>"]}' \
  "$API/api/realms/public/users/$ID/membership"                            # roles and groups
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"mode":"link"}' "$API/api/realms/public/users/$ID/credential"       # password reset
curl -s -X POST -H "$H" "$API/api/realms/public/users/$ID/logout"          # end sessions
curl -s -X DELETE -H "$H" "$API/api/realms/public/users/$ID"               # delete
```

Available roles and groups: `GET /api/realms/public/catalog`.

### Applications and access

Joining the realm grants access to no application: access is the user's roles in each
application. The application is a resource client (audience and roles); what users log in to are the web apps
linked to it (`application` in `PUT .../web-apps/{clientId}`). Those marked `declared: true` come from
Terraform and are read-only.

```bash
curl -s -H "$H" "$API/api/realms/public/applications"                       # applications, roles, clients, theme
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"id":"notes","name":"Notes","roles":["reader","editor"]}' \
  "$API/api/realms/public/applications"                                     # register
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"name":"admin"}' "$API/api/realms/public/applications/notes/roles"   # new role
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"theme":"acme-v3"}' "$API/api/realms/public/applications/notes/theme" # theme; empty falls back to the realm's
curl -s -H "$H" "$API/api/realms/public/users/$ID/access"                   # user access
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"roles":["reader"]}' "$API/api/realms/public/users/$ID/access/notes" # set; [] revokes
```

### Operators and API clients

Operators are the users of `private`. `private` only lets in those registered beforehand, with
their Google Workspace email; the account is linked at first login.

```bash
curl -s -H "$H" "$API/api/operators?q=ana"
curl -s -X POST -H "$H" -H 'Content-Type: application/json' "$API/api/operators" -d '{
  "email":"ana@criarch.com.br","firstName":"Ana","lastName":"Souza",
  "assignment":{"global":["iam-auditor"],"realms":{"public":"viewer"}}
}'
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"global":[],"realms":{"public":"admin"}}' "$API/api/operators/$ID/roles"   # complete state
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"name":"Pipeline","assignment":{"global":[],"realms":{"public":"admin"}}}' \
  "$API/api/api-clients"                     # returns the secret ONCE
```

### A realm's API keys (client_credentials for applications)

```bash
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"name":"Billing worker","description":"exports invoices","roles":["billing-reader"]}' \
  "$API/api/realms/public/api-keys"            # returns the secret ONCE; keep it in a vault
curl -s -X POST -H "$H" "$API/api/realms/public/api-keys/$KEY_ID/rotate"
```

### Events and audit

```bash
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"eventsEnabled":true,"adminEventsEnabled":true,"eventsExpirationDays":30}' \
  "$API/api/realms/public/events/config"
curl -s -H "$H" "$API/api/realms/public/events?max=100"
curl -s -H "$H" "$API/api/audit?limit=50"
```

### Themes

```bash
curl -s -X POST -H "$H" -F family=acme -F file=@theme.zip "$API/api/themes"
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"login":"acme-v1"}' "$API/api/realms/public/themes"            # realm theme
```

## 5. Client examples

### Python

```python
import os, time, requests

class IAM:
    def __init__(self):
        self.api = os.environ["IAM_API_URL"]
        self.token_url = os.environ["IAM_TOKEN_URL"]
        self.auth = (os.environ["IAM_CLIENT_ID"], os.environ["IAM_CLIENT_SECRET"])
        self._token, self._expires = None, 0

    def token(self):
        if time.time() > self._expires - 30:
            r = requests.post(self.token_url, data={"grant_type": "client_credentials"}, auth=self.auth, timeout=10)
            r.raise_for_status()
            body = r.json()
            self._token, self._expires = body["access_token"], time.time() + body["expires_in"]
        return self._token

    def call(self, method, path, **kwargs):
        r = requests.request(method, self.api + path, headers={"Authorization": f"Bearer {self.token()}"}, timeout=30, **kwargs)
        if r.status_code >= 400:
            raise RuntimeError(f"{r.status_code}: {r.json().get('error')}")
        return r.json() if r.content else None

iam = IAM()
print(iam.call("GET", "/api/realms"))
```

### TypeScript

```ts
const api = process.env.IAM_API_URL!;
let cached = { token: "", expires: 0 };

async function token() {
  if (Date.now() < cached.expires - 30_000) return cached.token;
  const body = new URLSearchParams({ grant_type: "client_credentials", client_id: process.env.IAM_CLIENT_ID!, client_secret: process.env.IAM_CLIENT_SECRET! });
  const r = await fetch(process.env.IAM_TOKEN_URL!, { method: "POST", body });
  const json = await r.json();
  cached = { token: json.access_token, expires: Date.now() + json.expires_in * 1000 };
  return cached.token;
}

export async function iam<T>(method: string, path: string, payload?: unknown): Promise<T> {
  const r = await fetch(api + path, {
    method,
    headers: { Authorization: `Bearer ${await token()}`, ...(payload ? { "Content-Type": "application/json" } : {}) },
    body: payload ? JSON.stringify(payload) : undefined,
  });
  if (!r.ok) throw new Error(`${r.status}: ${(await r.json()).error}`);
  return (r.status === 204 ? undefined : await r.json()) as T;
}
```

## 6. Rules for agents

- Read `GET /api/me` first and do not attempt what the permissions do not cover.
- Before deleting (user, API key, API client, web app), confirm with whoever asked. Deletion cannot be undone.
- Prefer `credential: "link"` to create accounts: the link goes to the person, never to the agent.
- Returned secrets (API keys, temporary passwords, links) only appear once; hand them to the
  requester through a secure channel and do not repeat them afterwards.
- On 401, renew the token once; if it persists, stop and report.
