import ast
import sys
from pathlib import Path

PACKAGE = "sample_app"
SOURCE = Path(__file__).resolve().parents[2] / "src" / PACKAGE

ALLOWED = {
    "domain": {"domain"},
    "application": {"domain", "application"},
    "adapters": {"domain", "application", "adapters"},
}
STANDARD_LIBRARY_ONLY = {"domain", "application"}


def module_parts(source: Path, path: Path) -> list[str]:
    parts = list(path.relative_to(source.parent).with_suffix("").parts)
    if parts[-1] == "__init__":
        parts.pop()
    return parts


def imported_modules(source: Path, path: Path) -> list[str]:
    tree = ast.parse(path.read_text(encoding="utf-8"), filename=str(path))
    package = module_parts(source, path)
    if path.name != "__init__.py":
        package = package[:-1]
    found = []
    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            found.extend(alias.name for alias in node.names)
        if isinstance(node, ast.ImportFrom):
            base = package[: len(package) - node.level + 1] if node.level else []
            found.append(".".join([*base, *([node.module] if node.module else [])]))
    return found


def violations(source: Path) -> list[str]:
    errors = []
    for path in sorted(source.rglob("*.py")):
        layer = path.relative_to(source).parts[0]
        if layer not in ALLOWED:
            continue
        module = ".".join(module_parts(source, path))
        for imported in imported_modules(source, path):
            root, _, rest = imported.partition(".")
            if root == source.name:
                if rest.split(".")[0] not in ALLOWED[layer]:
                    errors.append(f"{module}: {layer} must not depend on {imported}")
                continue
            if layer in STANDARD_LIBRARY_ONLY and root not in sys.stdlib_module_names:
                errors.append(f"{module}: {layer} must not depend on {imported}")
    return errors


def test_dependencies_point_inward() -> None:
    assert violations(SOURCE) == []


def test_outward_dependencies_are_reported(tmp_path: Path) -> None:
    source = tmp_path / PACKAGE
    (source / "domain").mkdir(parents=True)
    (source / "application").mkdir()
    (source / "domain" / "model.py").write_text("import fastapi\nfrom datetime import datetime\n")
    (source / "application" / "service.py").write_text("from ..adapters.http import app\n")

    assert violations(source) == [
        f"{PACKAGE}.application.service: application must not depend on {PACKAGE}.adapters.http",
        f"{PACKAGE}.domain.model: domain must not depend on fastapi",
    ]
