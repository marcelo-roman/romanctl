# SampleApp

API Python em camadas com Django, Django REST Framework e uv, gerada pelo `romanctl`.

```bash
romanctl run sample-app        # aplica as migrações e sobe o servidor
uv run pytest
uv run ruff check
uv run python manage.py makemigrations
```

Endereço: http://localhost:8001. A porta muda com a variável `PORT`. O `uv run sample-app` roda `migrate` e depois
`runserver`, e o `manage.py` serve para o resto.

## Camadas

Cada app do Django é dividido em camadas, e as dependências descem sempre: views → services e selectors → models. O
teste em `tests/test_layers.py` garante isso e também que as views não usam o ORM direto.

| Arquivo | Conteúdo | Depende de |
|---|---|---|
| `tasks/models.py` | Modelos do Django | nada do projeto |
| `tasks/selectors.py` | Consultas | models |
| `tasks/services.py` | Regras de negócio e escritas | models, selectors |
| `tasks/serializers.py` | Validação da entrada e formato da saída | models |
| `tasks/views.py` | Views finas do DRF: validam, chamam services ou selectors e respondem | serializers, services, selectors |
| `tasks/urls.py` | Rotas do app | views |
| `sample_app/` | Settings, URLs raiz, WSGI/ASGI e o tratador de erros | |

## Endpoints

O JSON usa snake_case, como os campos dos modelos.

| Método | Caminho | O que faz |
|---|---|---|
| `POST` | `/tasks` | Cria uma tarefa: `{"title": "Comprar pão", "due_date": "2030-01-31"}` |
| `GET` | `/tasks?done=false&limit=50` | As tarefas mais recentes; `limit` vai de 1 a 200, 50 por padrão |
| `GET` | `/tasks/{id}` | Uma tarefa |
| `PATCH` | `/tasks/{id}` | Muda `title`, `done` ou `due_date` |
| `DELETE` | `/tasks/{id}` | Remove a tarefa |
| `GET` | `/health` | Saúde da API |

Uma tarefa tem `id`, `title` (obrigatório, até 200 caracteres, sem espaços nas pontas), `done`, `due_date` (opcional,
nunca no passado ao criar ou mudar o prazo) e `created_at`.

Corpos inválidos respondem 400, tarefas que não existem, 404, e regras de negócio violadas, 422, sempre como
`application/problem+json`: `{"status": 422, "title": "Regra de negócio violada", "detail": "..."}`.
Rotas que não existem também respondem assim, salvo com `DJANGO_DEBUG=true`, o padrão em desenvolvimento, em que
o Django mostra a página de depuração.

## Persistência

PostgreSQL com psycopg 3, lido de `DATABASE_URL`. O `romanctl up sample-app` cria o banco `sample_app` e grava a
variável no `.env`. Sem `DATABASE_URL`, como no `romanctl run`, que não lê o `.env`, o projeto usa SQLite em
`db.sqlite3`, que fica fora do git; o caminho muda com `SQLITE_PATH`.

As tabelas vêm das migrações do Django, em `tasks/migrations`. Ao mudar um modelo, rode
`uv run python manage.py makemigrations` e versione o arquivo gerado.

Os testes rodam em SQLite em memória. Com `DATABASE_URL` definida, a suíte inteira roda no PostgreSQL, num banco de
teste que o pytest-django cria e apaga, e os testes de `tests/postgres` também rodam.

## Produção

`DJANGO_DEBUG=false` exige `DJANGO_SECRET_KEY`. `DJANGO_ALLOWED_HOSTS` recebe os domínios separados por vírgula.
