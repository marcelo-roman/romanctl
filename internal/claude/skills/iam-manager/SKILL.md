---
name: iam-manager
description: Administra o Keycloak pela API REST do IAM Manager com uma credencial client_credentials — criar, buscar, bloquear e excluir usuários, gerar links de convite e reset de senha, cadastrar aplicações e dar acesso a elas, cadastrar operadores, criar e rotacionar API keys, ler eventos de login e a auditoria, aplicar temas. Use quando o pedido envolver usuários, senhas, aplicações, permissões, API keys ou login de algum app que usa o Keycloak, ou quando aparecerem IAM_CLIENT_ID e IAM_CLIENT_SECRET.
---

# IAM Manager

O IAM Manager é a porta de administração do Keycloak. Tudo passa pela API dele, nunca pelo console
admin do Keycloak nem pela Admin API direto: ela valida permissões, protege o realm `master` e
registra cada alteração na auditoria com o seu `client_id`.

São dois realms: `private` (operadores e ferramentas internas, login só pelo Google Workspace da
criarch) e `public` (usuários das aplicações). Não há como criar ou excluir realms pela API.

A fonte da verdade é a especificação: `GET $IAM_API_URL/api/openapi.json`. O guia completo, com
receitas, está em `GET $IAM_API_URL/api/guide.md` e em [references/api.md](references/api.md).

## Credencial

A credencial é um **cliente da API** (client_credentials) e vem por variáveis de ambiente:

| variável | exemplo |
| --- | --- |
| `IAM_API_URL` | `https://iam-api.marceloroman.com.br` |
| `IAM_TOKEN_URL` | `https://auth.marceloroman.com.br/realms/private/protocol/openid-connect/token` |
| `IAM_CLIENT_ID` | `iam-api-...` |
| `IAM_CLIENT_SECRET` | segredo |

Se elas não estiverem no ambiente, procure `~/.config/iam-manager/credentials.env` e carregue com
`set -a; . ~/.config/iam-manager/credentials.env; set +a`. Se não houver credencial, peça ao
usuário; não invente nem procure em outros lugares.

Com o `romanctl` instalado, `romanctl iam` faz as mesmas chamadas com essa credencial e cuida do
token: `romanctl iam credentials --json` diz se ela existe, e `romanctl help iam` lista os comandos.
Para gravar, o usuário roda `! romanctl iam credentials set --client-id <id>`, que pede o secret
sem eco.

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

`globalPermissions` vale para todos os realms; `realmPermissions` lista os realms delegados
(`public:admin` ou `public:viewer`). Não
tente o que as permissões não cobrem: um 403 é resposta final, não algo para contornar.

## Regras

- **Exclusão não tem volta.** Antes de excluir usuário, API key, cliente da API, web app ou versão
  de tema, confirme com o usuário dizendo exatamente o que vai sumir.
- **Contas novas com `credential: "link"`.** O link vai para a pessoa, não fica com você. Senha
  temporária só se o usuário pedir.
- **Usuários precisam de** nome, sobrenome e e-mail no perfil, senão o login trava em
  "Account is not fully set up".
- **Entrar no realm não dá acesso a aplicação nenhuma.** O acesso são as roles do usuário em cada
  aplicação (`PUT .../users/{id}/access/{app}`); lista vazia revoga.
- **Operador novo é cadastrado antes do primeiro login** (`POST /api/operators` com o e-mail do
  Workspace); o `private` recusa quem não foi cadastrado.
- **PUT de papéis/grupos (`/membership`) e de permissões recebe o estado desejado completo.** Leia o
  atual, altere e envie tudo; não mande só o que muda.
- **Campos desconhecidos são rejeitados (422).** Envie só o que a especificação define.
- **Listas de usuários são paginadas** (`first`, `max` até 100; a resposta traz `total`).
- **Secrets aparecem uma vez** (API key, rotação, senha temporária, link). Entregue ao usuário e não
  guarde nem repita depois.
- O realm `master` nunca é acessível. O `private` contém os operadores e só aceita `iam-admin`; só
  mexa nele se o usuário pedir explicitamente.

## Mapa rápido

| para | chamada |
| --- | --- |
| realms e configurações (só leitura) | `GET /api/realms`, `GET /api/realms/{realm}/settings` |
| usuários | `GET /api/realms/{realm}/users?search=&first=&max=`, `POST`, `GET/PUT/DELETE .../users/{id}` |
| papéis e grupos do usuário | `GET /api/realms/{realm}/catalog`, `PUT .../users/{id}/membership` |
| reset de senha | `POST .../users/{id}/credential` com `{"mode":"link"}` |
| aplicações | `GET` / `POST /api/realms/{realm}/applications`, `POST .../{app}/roles`, `PUT .../{app}/theme` |
| acesso do usuário às aplicações | `GET .../users/{id}/access`, `PUT .../users/{id}/access/{app}` com `{"roles":[...]}` |
| web apps (clients com login) | `GET /api/realms/{realm}/web-apps`, `PUT/DELETE .../web-apps/{clientId}` |
| operadores | `GET /api/operators?q=`, `POST /api/operators`, `PUT /api/operators/{id}/roles` |
| clientes da API (este tipo de credencial) | `GET/POST /api/api-clients`, `PUT/DELETE .../{id}`, `POST .../{id}/rotate` |
| API keys de um app | `GET/POST /api/realms/{realm}/api-keys`, `POST .../{id}/rotate` |
| eventos de login | `GET /api/realms/{realm}/events`; ligar: `PUT .../events/config` |
| auditoria | `GET /api/audit?limit=` |
| temas | `POST /api/themes` (multipart), `PUT /api/realms/{realm}/themes` |

Exemplos completos de corpo para cada caso em [references/api.md](references/api.md).

## Erros

Sempre `{"error":"..."}`: 401 token, 403 sem permissão, 404 não existe, 409 conflito (nome já
existe, recurso em uso), 422 validação (leia a mensagem), 429 espere um minuto, 503 Keycloak fora —
tente de novo com espera.
