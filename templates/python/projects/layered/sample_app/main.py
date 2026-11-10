import os
import sys

from django.core.management import execute_from_command_line


def run() -> None:
    os.environ.setdefault("DJANGO_SETTINGS_MODULE", "sample_app.settings")
    address = f"{os.environ.get('HOST', '127.0.0.1')}:{os.environ.get('PORT', '8001')}"
    execute_from_command_line([sys.argv[0], "migrate", "--no-input"])
    execute_from_command_line([sys.argv[0], "runserver", address])
