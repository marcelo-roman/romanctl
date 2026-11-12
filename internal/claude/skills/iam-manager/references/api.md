# IAM Manager API — guia de uso

API REST para administrar os realms `private` (operadores e ferramentas internas) e `public`
(usuários das aplicações) do Keycloak: usuários, credenciais, aplicações e acesso, API keys, login
social, web apps, eventos, temas, operadores e auditoria. Este guia serve para pessoas e para agentes de IA.

- Especificação completa: `GET {API}/api/openapi.json` (ou `.yaml`)
- Este guia: `GET {API}/api/guide.md`

## 1. Endereços

| nome | valor em produção |
| --- | --- |
| `API` | `https://iam-api.marceloroman.com.br` |
| `TOKEN_URL` | `https://auth.marceloroman.com.br/realms/private/protocol/openid-connect/token` |

## 2. Autenticação (client_credentials)

Você recebe um `client_id` e um `client_secret` de um **cliente da API** (criado em
IAM Manager → Operadores → Clientes da API, no realm `private`). Troque-os por um access token
com audiência `iam-manager-api`:

```bash
TOKEN=$(curl -s -X POST "$TOKEN_URL" \
  -d grant_type=client_credentials \
  -d client_id="$IAM_CLIENT_ID" \
  -d client_secret="$IAM_CLIENT_SECRET" | jq -r .access_token)
```

- O token vale 5 minutos (`expires_in`). Reaproveite até perto de expirar e peça outro; não há
  refresh token no client_credentials.
- Envie em toda chamada: `Authorization: Bearer $TOKEN`.
- **Nunca** registre o secret ou o token em log, commit, issue ou resposta de chat.

Confira quem você é e o que pode fazer:

```bash
curl -s -H "Authorization: Bearer $TOKEN" "$API/api/me"
```

`globalPermissions` vale para todos os realms; `realmPermissions` lista realms delegados. Os papéis
são client roles de `iam-manager-api`: globais `iam-admin`, `iam-user-admin`, `iam-app-admin`,
`iam-theme-admin` e `iam-auditor`, e as delegações `public:admin` e `public:viewer`.

## 3. Convenções

- JSON em tudo (`Content-Type: application/json`), exceto o envio de tema (`multipart/form-data`).
- Campos desconhecidos no corpo são rejeitados (422). Envie só o que a especificação define.
- Listas de usuários são paginadas: `first` (deslocamento) e `max` (até 100); a resposta traz `total`.
- Os PUT de papéis/grupos e de permissões recebem o **estado desejado completo**, não um diff.
- `realm` é `private` ou `public`; não há como criar ou excluir realms pela API.
- O realm `master` nunca é acessível. O `private` (operadores) só com `iam-admin`.
- Toda alteração fica na auditoria (`GET /api/audit`) com o `client_id` como autor.

### Erros

Corpo sempre `{"error": "mensagem"}`.

| status | significa | o que fazer |
| --- | --- | --- |
| 401 | token ausente, expirado ou inválido | pegue outro token |
| 403 | falta permissão para aquele realm/ação | não insista; reporte |
| 404 | realm, usuário ou recurso não existe | confira o identificador |
| 409 | conflito (nome já existe, versão em uso) | escolha outro nome ou resolva o uso |
| 422 | validação (campo, senha fraca, regra) | leia a mensagem e corrija |
| 429 | limite de tentativas | espere um minuto |
| 503 | Keycloak indisponível | tente de novo com espera |

## 4. Receitas

Variáveis usadas: `API`, `TOKEN`. Use `H='Authorization: Bearer '$TOKEN`.

### Realms

```bash
curl -s -H "$H" "$API/api/realms"                                       # listar
curl -s -H "$H" "$API/api/realms/public/settings"                       # configurações (só leitura)
```

As configurações vêm do Terraform; a API só as mostra.

### Usuários

```bash
# criar com link de convite (a pessoa define a senha)
curl -s -X POST -H "$H" -H 'Content-Type: application/json' "$API/api/realms/public/users" -d '{
  "profile": {"username":"ana","email":"ana@acme.com","firstName":"Ana","lastName":"Silva","enabled":true,"emailVerified":true},
  "credential": "link", "roles": [], "groups": []
}'
```

`credential`: `link` (devolve `credential.link`, uso único, 72 h), `temporary` (devolve
`credential.temporaryPassword`, troca obrigatória no primeiro login), `email` (o Keycloak envia; exige
SMTP no realm) ou `none`. O perfil precisa de nome, sobrenome e e-mail para o login funcionar.

```bash
curl -s -H "$H" "$API/api/realms/public/users?search=ana&first=0&max=25"   # buscar
curl -s -H "$H" "$API/api/realms/public/users/$ID"                         # detalhe
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"roles":["billing-reader"],"groups":["<id-do-grupo>"]}' \
  "$API/api/realms/public/users/$ID/membership"                            # papéis e grupos
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"mode":"link"}' "$API/api/realms/public/users/$ID/credential"       # reset de senha
curl -s -X POST -H "$H" "$API/api/realms/public/users/$ID/logout"          # encerrar sessões
curl -s -X DELETE -H "$H" "$API/api/realms/public/users/$ID"               # excluir
```

Papéis e grupos disponíveis: `GET /api/realms/public/catalog`.

### Aplicações e acesso

Entrar no realm não dá acesso a aplicação nenhuma: o acesso são as roles do usuário em cada
aplicação. A aplicação é um client de recurso (audiência e roles); quem faz login são as web apps
ligadas a ela (`application` no `PUT .../web-apps/{clientId}`). As marcadas `declared: true` vêm do
Terraform e são só leitura.

```bash
curl -s -H "$H" "$API/api/realms/public/applications"                       # aplicações, roles, clients, tema
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"id":"notes","name":"Notes","roles":["reader","editor"]}' \
  "$API/api/realms/public/applications"                                     # cadastrar
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"name":"admin"}' "$API/api/realms/public/applications/notes/roles"   # nova role
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"theme":"acme-v3"}' "$API/api/realms/public/applications/notes/theme" # tema; vazio volta ao do realm
curl -s -H "$H" "$API/api/realms/public/users/$ID/access"                   # acesso do usuário
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"roles":["reader"]}' "$API/api/realms/public/users/$ID/access/notes" # definir; [] revoga
```

### Operadores e clientes da API

Operadores são os usuários do `private`. O `private` só deixa entrar quem foi cadastrado antes, com
o e-mail do Google Workspace; a conta se vincula no primeiro login.

```bash
curl -s -H "$H" "$API/api/operators?q=ana"
curl -s -X POST -H "$H" -H 'Content-Type: application/json' "$API/api/operators" -d '{
  "email":"ana@criarch.com.br","firstName":"Ana","lastName":"Souza",
  "assignment":{"global":["iam-auditor"],"realms":{"public":"viewer"}}
}'
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"global":[],"realms":{"public":"admin"}}' "$API/api/operators/$ID/roles"   # estado completo
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"name":"Pipeline","assignment":{"global":[],"realms":{"public":"admin"}}}' \
  "$API/api/api-clients"                     # devolve secret UMA vez
```

### API keys de um realm (client_credentials para aplicações)

```bash
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"name":"Billing worker","description":"exporta faturas","roles":["billing-reader"]}' \
  "$API/api/realms/public/api-keys"            # devolve secret UMA vez; guarde em cofre
curl -s -X POST -H "$H" "$API/api/realms/public/api-keys/$KEY_ID/rotate"
```

### Eventos e auditoria

```bash
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"eventsEnabled":true,"adminEventsEnabled":true,"eventsExpirationDays":30}' \
  "$API/api/realms/public/events/config"
curl -s -H "$H" "$API/api/realms/public/events?max=100"
curl -s -H "$H" "$API/api/audit?limit=50"
```

### Temas

```bash
curl -s -X POST -H "$H" -F family=acme -F file=@tema.zip "$API/api/themes"
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"login":"acme-v1"}' "$API/api/realms/public/themes"            # tema do realm
```

## 5. Exemplos de cliente

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

## 6. Regras para agentes

- Leia `GET /api/me` primeiro e não tente o que as permissões não cobrem.
- Antes de excluir (usuário, API key, cliente da API, web app), confirme com quem pediu. Exclusão não tem volta.
- Prefira `credential: "link"` para criar contas: o link vai para a pessoa, nunca para o agente.
- Secrets devolvidos (API keys, senhas temporárias, links) só aparecem uma vez; entregue-os ao
  solicitante por canal seguro e não os repita depois.
- Em 401, renove o token uma vez; se continuar, pare e reporte.
