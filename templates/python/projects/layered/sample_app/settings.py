from pathlib import Path

from sample_app.database import default_database
from sample_app.environment import allowed_hosts, flag, secret_key

BASE_DIR = Path(__file__).resolve().parent.parent

DEBUG = flag("DJANGO_DEBUG", default=True)
SECRET_KEY = secret_key(DEBUG)
ALLOWED_HOSTS = allowed_hosts()

INSTALLED_APPS = [
    "rest_framework",
    "tasks",
    # romanctl:apps
]

MIDDLEWARE = [
    "django.middleware.security.SecurityMiddleware",
    "django.middleware.common.CommonMiddleware",
]

ROOT_URLCONF = "sample_app.urls"
WSGI_APPLICATION = "sample_app.wsgi.application"
ASGI_APPLICATION = "sample_app.asgi.application"

DATABASES = {"default": default_database(BASE_DIR)}
DEFAULT_AUTO_FIELD = "django.db.models.BigAutoField"

LANGUAGE_CODE = "pt-br"
TIME_ZONE = "America/Sao_Paulo"
USE_I18N = True
USE_TZ = True
APPEND_SLASH = False

REST_FRAMEWORK = {
    "DEFAULT_RENDERER_CLASSES": ["rest_framework.renderers.JSONRenderer"],
    "DEFAULT_PARSER_CLASSES": ["rest_framework.parsers.JSONParser"],
    "DEFAULT_AUTHENTICATION_CLASSES": [],
    "DEFAULT_PERMISSION_CLASSES": [],
    "UNAUTHENTICATED_USER": None,
    "EXCEPTION_HANDLER": "sample_app.problems.problem_exception_handler",
    "TEST_REQUEST_DEFAULT_FORMAT": "json",
}

LOGGING = {
    "version": 1,
    "disable_existing_loggers": False,
    "formatters": {"plain": {"format": "%(asctime)s %(levelname)s %(name)s %(message)s"}},
    "handlers": {"console": {"class": "logging.StreamHandler", "formatter": "plain"}},
    "root": {"handlers": ["console"], "level": "INFO"},
    "loggers": {"django": {"handlers": [], "level": "INFO"}},
}
