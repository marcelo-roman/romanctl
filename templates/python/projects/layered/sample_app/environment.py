import os

from django.core.exceptions import ImproperlyConfigured

DEVELOPMENT_SECRET_KEY = "django-insecure-development-only"


def flag(name: str, default: bool) -> bool:
    value = os.environ.get(name, "").strip().lower()
    if not value:
        return default
    return value in {"1", "true", "yes", "on"}


def secret_key(debug: bool) -> str:
    key = os.environ.get("DJANGO_SECRET_KEY", "")
    if key:
        return key
    if not debug:
        raise ImproperlyConfigured("Set DJANGO_SECRET_KEY when DJANGO_DEBUG is false.")
    return DEVELOPMENT_SECRET_KEY


def allowed_hosts() -> list[str]:
    hosts = os.environ.get("DJANGO_ALLOWED_HOSTS", "localhost,127.0.0.1")
    return [host.strip() for host in hosts.split(",") if host.strip()]
