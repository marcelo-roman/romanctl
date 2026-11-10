FROM ghcr.io/astral-sh/uv:python3.12-bookworm-slim AS build
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy UV_PYTHON_DOWNLOADS=never
WORKDIR /app
COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev --no-install-project
COPY . .
RUN uv sync --frozen --no-dev --no-editable

FROM python:3.12-slim-bookworm
WORKDIR /app
COPY --from=build /app/.venv /app/.venv
ENV PATH="/app/.venv/bin:$PATH" PORT=8080 DJANGO_SETTINGS_MODULE=[[.Module]].settings DJANGO_DEBUG=false \
    DJANGO_ALLOWED_HOSTS=* SQLITE_PATH=/app/db.sqlite3
EXPOSE 8080
CMD ["sh", "-c", "django-admin migrate --no-input && exec gunicorn [[.Module]].wsgi --bind 0.0.0.0:${PORT}"]
