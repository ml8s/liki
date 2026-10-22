#!/usr/bin/env python3
"""Validate the canonical domain language and reject unregistered alias drift."""


import argparse
import fnmatch
import json
import re
import sys
from pathlib import Path

try:
    import jsonschema
except ImportError:  # pragma: no cover - make check installs project dependencies
    jsonschema = None


ROOT = Path(__file__).resolve().parents[1]
SCHEMA = ROOT / "tools/domain-language.schema.json"
REGISTRY = ROOT / "tools/domain-language.json"

SCAN_EXTENSIONS = {".go", ".py", ".md", ".json", ".yaml", ".yml", ".csv"}
SKIP_DIRECTORIES = {
    ".git",
    ".venv",
    "__pycache__",
    "node_modules",
    "dist",
    "build",
    "htmlcov",
    ".pytest_cache",
    ".ruff_cache",
}
SKIP_FILES = {"CHANGELOG.md"}
SOURCE_FILES = {REGISTRY.resolve(), (ROOT / "docs/DOMAIN_LANGUAGE.md").resolve()}
SOURCE_FILES.add((ROOT / "docs/FUNCTION_FIRST_REFACTOR_PLAN.md").resolve())


def _candidate_files() -> list[Path]:
    files: list[Path] = []
    for path in ROOT.rglob("*"):
        if not path.is_file() or path.is_symlink():
            continue
        relative = path.relative_to(ROOT)
        if any(part in SKIP_DIRECTORIES for part in relative.parts) or path.name in SKIP_FILES:
            continue
        if path.resolve() in SOURCE_FILES or path.suffix not in SCAN_EXTENSIONS:
            continue
        files.append(path)
    return sorted(files)


def _alias_pattern(alias: str) -> re.Pattern[str]:
    return re.compile(r"(?<![A-Za-z0-9_])" + re.escape(alias) + r"(?![A-Za-z0-9_])")


def _scopes_for(concept: dict, alias: str) -> list[str]:
    scopes: list[str] = []
    for item in concept.get("compatibility", []):
        if item["name"] == alias:
            scopes.extend(item["scope"])
    return scopes


def _allowed(relative: str, scopes: list[str]) -> bool:
    return any(fnmatch.fnmatchcase(relative, scope) for scope in scopes)


def _validate_uniqueness(registry: dict) -> list[str]:
    errors: list[str] = []
    ids: set[str] = set()
    surfaces: dict[str, dict[str, str]] = {"go": {}, "python": {}, "json": {}, "zh_cn": {}}
    compatibility_names: dict[str, str] = {}

    for concept in registry["concepts"]:
        concept_id = concept["id"]
        if concept_id in ids:
            errors.append(f"duplicate concept id: {concept_id}")
        ids.add(concept_id)
        for surface, canonical in concept["surfaces"].items():
            key = canonical.lower()
            if key in surfaces[surface]:
                errors.append(
                    f"duplicate {surface} projection {canonical!r}: "
                    f"{surfaces[surface][key]} and {concept_id}"
                )
            surfaces[surface][key] = concept_id
        for item in concept.get("compatibility", []):
            name = item["name"]
            if name in compatibility_names and compatibility_names[name] != concept_id:
                errors.append(f"compatibility alias {name!r} belongs to multiple concepts")
            compatibility_names[name] = concept_id

    canonical_names = {value for by_surface in surfaces.values() for value in by_surface}
    for concept in registry["concepts"]:
        for item in concept.get("compatibility", []):
            if item["name"].lower() in canonical_names:
                errors.append(
                    f"{concept['id']}: compatibility name {item['name']!r} is also canonical"
                )
        for related in concept.get("related", []):
            if related not in ids:
                errors.append(f"{concept['id']}: unknown related concept {related!r}")
    return errors


def _scan_files(registry: dict) -> list[str]:
    errors: list[str] = []
    aliases: list[tuple[dict, str, re.Pattern[str]]] = []
    for concept in registry["concepts"]:
        names = set(concept.get("forbidden", []))
        names.update(item["name"] for item in concept.get("compatibility", []))
        for name in names:
            if any(name.lower() == canonical.lower() for canonical in concept["surfaces"].values()):
                errors.append(f"{concept['id']}: forbidden/compatibility name is canonical: {name}")
                continue
            aliases.append((concept, name, _alias_pattern(name)))

    for path in _candidate_files():
        relative = path.relative_to(ROOT).as_posix()
        try:
            text = path.read_text(encoding="utf-8")
        except (OSError, UnicodeError) as error:
            errors.append(f"{relative}: unreadable scan target: {error}")
            continue
        for concept, alias, pattern in aliases:
            match = pattern.search(text)
            if match is None or _allowed(relative, _scopes_for(concept, alias)):
                continue
            line = text.count("\n", 0, match.start()) + 1
            errors.append(
                f"{relative}:{line}: unregistered alias {alias!r}; "
                f"use {concept['surfaces']['json']!r} for concept {concept['id']!r}"
            )
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.parse_args()
    try:
        registry = json.loads(REGISTRY.read_text(encoding="utf-8"))
        schema = json.loads(SCHEMA.read_text(encoding="utf-8"))
        if jsonschema is None:
            raise RuntimeError("jsonschema is required to validate tools/domain-language.json")
        jsonschema.Draft202012Validator.check_schema(schema)
        jsonschema.validate(registry, schema, cls=jsonschema.Draft202012Validator)
    except Exception as error:
        print(f"domain language registry is invalid: {error}", file=sys.stderr)
        return 1

    errors = _validate_uniqueness(registry)
    errors.extend(_scan_files(registry))

    count = len(registry["concepts"])
    if errors:
        print(f"domain language violations ({len(errors)}):", file=sys.stderr)
        for error in errors:
            print(f"  ✗ {error}", file=sys.stderr)
        return 1
    print(f"[check-domain-language] ✓ {count} concepts, aliases guarded")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
