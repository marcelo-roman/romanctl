---
name: iam-manager
description: Administra o Keycloak pela API REST do IAM Manager com uma credencial client_credentials — criar e configurar realms, criar, buscar, bloquear e excluir usuários, gerar links de convite e reset de senha, colocar a mesma pessoa em vários realms, criar e rotacionar API keys, ler eventos de login e a auditoria, aplicar temas. Use quando o pedido envolver realms, usuários, senhas, permissões, API keys ou login de algum app que usa o Keycloak, ou quando aparecerem IAM_CLIENT_ID e IAM_CLIENT_SECRET.
---

# IAM Manager

O IAM Manager é a porta de administração do Keycloak. Tudo passa pela API dele, nunca pelo console
admin do Keycloak nem pela Admin API direto: ela valida permissões, protege o realm `master` e
registra cada alteração na auditoria com o seu `client_id`.

A fonte da verdade é a especificação: `GET $IAM_API_URL/api/openapi.json`. O guia completo, com
receitas, está em `GET $IAM_API_URL/api/guide.md` e em [references/api.md](references/api.md).

## Credencial

A credencial é um **cliente da API** (client_credentials) e vem por variáveis de ambiente:

| variável | exemplo |
| --- | --- |
| `IAM_API_URL` | `https://iam-api.marceloroman.com.br` |
| `IAM_TOKEN_URL` | `https://auth.marceloroman.com.br/realms/iam-manager/protocol/openid-connect/token` |
| `IAM_CLIENT_ID` | `iam-api-...` |
| `IAM_CLIENT_SECRET` | segredo |

Se elas não estiverem no ambiente, procure `~/.config/iam-manager/credentials.env` e carregue com
`set -a; . ~/.config/iam-manager/credentials.env; set +a`. Se não houver credencial, peça ao
usuário; não invente nem procure em outros lugares.

**Nunca** mostre, registre, commite ou repita o `IAM_CLIENT_SECRET`, o token, senhas temporárias ou
links de convite em resposta, log ou arquivo. Para guardar a credencial, use o arquivo acima com
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

O token vale 5 minutos. Em 401, gere outro **uma vez**; se continuar, pare e reporte.

## Antes de qualquer coisa

```bash
iam "$IAM_API_URL/api/me"
```

`globalPermissions` vale para todos os realms; `realmPermissions` lista os realms delegados. Não
tente o que as permissões não cobrem: um 403 é resposta final, não algo para contornar.

## Regras

- **Exclusão não tem volta.** Antes de excluir realm, usuário, identidade, API key ou versão de tema,
  confirme com o usuário dizendo exatamente o que vai sumir. Excluir realm exige
  `{"confirm":"<realm>"}` no corpo.
- **Contas novas com `credential: "link"`.** O link vai para a pessoa, não fica com você. Senha
  temporária só se o usuário pedir.
- **Realms novos exigem** nome, sobrenome e e-mail no perfil do usuário, senão o login trava em
  "Account is not fully set up".
- **PUT de papéis/grupos (`/membership`) e de permissões recebe o estado desejado completo.** Leia o
  atual, altere e envie tudo; não mande só o que muda.
- **Campos desconhecidos são rejeitados (422).** Envie só o que a especificação define.
- **Listas de usuários são paginadas** (`first`, `max` até 100; a resposta traz `total`).
- **Secrets aparecem uma vez** (API key, rotação, senha temporária, link). Entregue ao usuário e não
  guarde nem repita depois.
- O realm `master` nunca é acessível. O realm `iam-manager` contém os operadores; só mexa nele se o
  usuário pedir explicitamente.

## Mapa rápido

| para | chamada |
| --- | --- |
| listar / criar realm | `GET` / `POST /api/realms` |
| configurar realm (SMTP, auto-cadastro…) | `GET` / `PUT /api/realms/{realm}/settings` |
| usuários | `GET /api/realms/{realm}/users?search=&first=&max=`, `POST`, `GET/PUT/DELETE .../users/{id}` |
| papéis e grupos do usuário | `GET /api/realms/{realm}/catalog`, `PUT .../users/{id}/membership` |
| reset de senha | `POST .../users/{id}/credential` com `{"mode":"link"}` |
| mesma pessoa em outro realm | `POST /api/identities/link`; buscar: `GET /api/identities?q=` |
| ação em todas as contas da pessoa | `POST /api/identities/{globalId}/actions` |
| API keys de um app | `GET/POST /api/realms/{realm}/api-keys`, `POST .../{id}/rotate` |
| eventos de login | `GET /api/realms/{realm}/events`; ligar: `PUT .../events/config` |
| auditoria | `GET /api/audit?limit=` |
| temas | `POST /api/themes` (multipart), `PUT /api/realms/{realm}/themes` |

Exemplos completos de corpo para cada caso em [references/api.md](references/api.md).

## Erros

Sempre `{"error":"..."}`: 401 token, 403 sem permissão, 404 não existe, 409 conflito (nome já
existe, recurso em uso), 422 validação (leia a mensagem), 429 espere um minuto, 503 Keycloak fora —
tente de novo com espera.
