import ast
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

ALLOWED = {
    "urls": {"views"},
    "views": {"serializers", "services", "selectors"},
    "serializers": {"models"},
    "services": {"models", "selectors"},
    "selectors": {"models"},
    "models": set(),
}
ORM_IN_VIEWS = {"objects", "save", "delete", "bulk_create", "bulk_update"}


def app_names(root: Path) -> set[str]:
    return {path.parent.name for path in root.glob("*/apps.py")}


def imported_modules(tree: ast.Module, package: str) -> list[str]:
    found = []
    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            found.extend(alias.name for alias in node.names)
        if isinstance(node, ast.ImportFrom):
            module = node.module or ""
            if node.level:
                module = ".".join(part for part in [package, module] if part)
            found.extend(f"{module}.{alias.name}" for alias in node.names)
    return found


def orm_calls(tree: ast.Module) -> list[str]:
    attributes = {node.attr for node in ast.walk(tree) if isinstance(node, ast.Attribute)}
    return sorted(attributes & ORM_IN_VIEWS)


def violations(root: Path) -> list[str]:
    apps = app_names(root)
    errors = []
    for app in sorted(apps):
        for layer in ALLOWED:
            path = root / app / f"{layer}.py"
            if not path.exists():
                continue
            tree = ast.parse(path.read_text(encoding="utf-8"), filename=str(path))
            for imported in imported_modules(tree, app):
                parts = imported.split(".")
                if parts[0] not in apps or len(parts) < 2 or parts[1] not in ALLOWED:
                    continue
                if parts[1] not in ALLOWED[layer]:
                    errors.append(f"{app}.{layer}: {layer} não pode importar {'.'.join(parts[:2])}")
            if layer == "views":
                errors.extend(f"{app}.views: views não podem usar o ORM ({call})" for call in orm_calls(tree))
    return errors


def test_layers_only_depend_downwards() -> None:
    assert violations(ROOT) == []


def test_violations_are_reported(tmp_path: Path) -> None:
    app = tmp_path / "notes"
    app.mkdir()
    (app / "apps.py").write_text("")
    (app / "views.py").write_text(
        "from notes.models import Note\nfrom .services import create\n\n"
        "def view():\n    Note.objects.all()\n    note = Note()\n    note.save()\n"
    )
    (app / "selectors.py").write_text("from notes import services\n")
    (app / "models.py").write_text("from notes.selectors import find\n")

    assert violations(tmp_path) == [
        "notes.views: views não pode importar notes.models",
        "notes.views: views não podem usar o ORM (objects)",
        "notes.views: views não podem usar o ORM (save)",
        "notes.selectors: selectors não pode importar notes.services",
        "notes.models: models não pode importar notes.selectors",
    ]
