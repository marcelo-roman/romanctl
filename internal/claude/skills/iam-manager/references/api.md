# IAM Manager API — guia de uso

API REST para administrar o Keycloak: realms, usuários, credenciais, identidades em vários realms,
API keys, eventos, temas e auditoria. Este guia serve para pessoas e para agentes de IA.

- Especificação completa: `GET {API}/api/openapi.json` (ou `.yaml`)
- Este guia: `GET {API}/api/guide.md`

## 1. Endereços

| nome | valor em produção |
| --- | --- |
| `API` | `https://iam-api.marceloroman.com.br` |
| `TOKEN_URL` | `https://auth.marceloroman.com.br/realms/iam-manager/protocol/openid-connect/token` |

## 2. Autenticação (client_credentials)

Você recebe um `client_id` e um `client_secret` de um **cliente da API** (criado em
IAM Manager → Operadores → Clientes da API). Troque-os por um access token:

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

`globalPermissions` vale para todos os realms; `realmPermissions` lista realms delegados.

## 3. Convenções

- JSON em tudo (`Content-Type: application/json`), exceto o envio de tema (`multipart/form-data`).
- Campos desconhecidos no corpo são rejeitados (422). Envie só o que a especificação define.
- Listas de usuários são paginadas: `first` (deslocamento) e `max` (até 100); a resposta traz `total`.
- Os PUT de papéis/grupos e de permissões recebem o **estado desejado completo**, não um diff.
- `realm` é o identificador (ex.: `acme`), não o nome de exibição.
- O realm `master` nunca é acessível. O realm `iam-manager` (operadores) só com `iam-admin`.
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
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"name":"acme","displayName":"ACME Corp"}' "$API/api/realms"      # criar
curl -s -H "$H" "$API/api/realms/acme/settings"                         # configurações
curl -s -X DELETE -H "$H" -H 'Content-Type: application/json' \
  -d '{"confirm":"acme"}' "$API/api/realms/acme"                        # excluir (irreversível)
```

O identificador do realm: minúsculas, números e hífen, 2–41 caracteres.

### Usuários

```bash
# criar com link de convite (a pessoa define a senha)
curl -s -X POST -H "$H" -H 'Content-Type: application/json' "$API/api/realms/acme/users" -d '{
  "profile": {"username":"ana","email":"ana@acme.com","firstName":"Ana","lastName":"Silva","enabled":true,"emailVerified":true},
  "credential": "link", "roles": [], "groups": []
}'
```

`credential`: `link` (devolve `credential.link`, uso único, 72 h), `temporary` (devolve
`credential.temporaryPassword`, troca obrigatória no primeiro login), `email` (o Keycloak envia; exige
SMTP no realm) ou `none`. Realms novos exigem nome, sobrenome e e-mail no perfil para o login funcionar.

```bash
curl -s -H "$H" "$API/api/realms/acme/users?search=ana&first=0&max=25"   # buscar
curl -s -H "$H" "$API/api/realms/acme/users/$ID"                         # detalhe
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"roles":["billing-reader"],"groups":["<id-do-grupo>"]}' \
  "$API/api/realms/acme/users/$ID/membership"                            # papéis e grupos
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"mode":"link"}' "$API/api/realms/acme/users/$ID/credential"       # reset de senha
curl -s -X POST -H "$H" "$API/api/realms/acme/users/$ID/logout"          # encerrar sessões
curl -s -X DELETE -H "$H" "$API/api/realms/acme/users/$ID"               # excluir
```

Papéis e grupos disponíveis: `GET /api/realms/acme/catalog`.

### Mesma pessoa em vários realms

```bash
curl -s -X POST -H "$H" -H 'Content-Type: application/json' "$API/api/identities/link" -d '{
  "sourceRealm":"acme","sourceUserId":"<id>","targetRealm":"beta","credential":"none"
}'
curl -s -H "$H" "$API/api/identities?q=ana"                              # buscar em todos os realms
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"action":"disable"}' "$API/api/identities/$GLOBAL_ID/actions"     # enable|disable|logout|delete|reset
```

### API keys de um realm (client_credentials para aplicações)

```bash
curl -s -X POST -H "$H" -H 'Content-Type: application/json' \
  -d '{"name":"Billing worker","description":"exporta faturas","roles":["billing-reader"]}' \
  "$API/api/realms/acme/api-keys"            # devolve secret UMA vez; guarde em cofre
curl -s -X POST -H "$H" "$API/api/realms/acme/api-keys/$KEY_ID/rotate"
```

### Eventos e auditoria

```bash
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"eventsEnabled":true,"adminEventsEnabled":true,"eventsExpirationDays":30}' \
  "$API/api/realms/acme/events/config"
curl -s -H "$H" "$API/api/realms/acme/events?max=100"
curl -s -H "$H" "$API/api/audit?limit=50"
```

### Temas

```bash
curl -s -X POST -H "$H" -F family=acme -F file=@tema.zip "$API/api/themes"
curl -s -X PUT -H "$H" -H 'Content-Type: application/json' \
  -d '{"login":"acme-v1"}' "$API/api/realms/acme/themes"
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
- Antes de excluir (realm, usuário, identidade, API key), confirme com quem pediu. Exclusão não tem volta.
- Prefira `credential: "link"` para criar contas: o link vai para a pessoa, nunca para o agente.
- Secrets devolvidos (API keys, senhas temporárias, links) só aparecem uma vez; entregue-os ao
  solicitante por canal seguro e não os repita depois.
- Em 401, renove o token uma vez; se continuar, pare e reporte.
