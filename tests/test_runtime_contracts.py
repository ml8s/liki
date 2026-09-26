from __future__ import annotations

import re
from pathlib import Path


ROOT = Path(__file__).parents[1]


def template_vars(path: Path) -> set[str]:
    return set(re.findall(r"^([A-Z][A-Z0-9_]+)=", path.read_text(encoding="utf-8"), re.MULTILINE))


def source_vars(paths, pattern: str) -> set[str]:
    found: set[str] = set()
    for path in paths:
        if path.name.endswith("_test.go"):
            continue
        found.update(re.findall(pattern, path.read_text(encoding="utf-8")))
    return found


def test_engine_runtime_environment_is_documented():
    go_files = list((ROOT / "engine/cmd").rglob("*.go")) + list(
        (ROOT / "engine/internal").rglob("*.go")
    )
    source = source_vars(
        go_files,
        r'\b(?:os\.Getenv|envOr)\(\s*"([A-Z][A-Z0-9_]+)"',
    )
    template = template_vars(ROOT / "engine/.env.example")
    assert not source - template

    runtime_doc = (ROOT / "docs/RUNTIME.md").read_text(encoding="utf-8")
    for name in source:
        assert f"`{name}`" in runtime_doc


def test_counsel_runtime_environment_is_documented():
    source = source_vars(
        (ROOT / "counsel/app").rglob("*.py"),
        r'\bos\.environ\.get\(\s*"([A-Z][A-Z0-9_]+)"',
    )
    template = template_vars(ROOT / "counsel/.env.example")
    assert not source - template

    runtime_doc = (ROOT / "docs/RUNTIME.md").read_text(encoding="utf-8")
    for name in source:
        assert f"`{name}`" in runtime_doc


def test_engine_image_carries_both_active_runtimes():
    dockerfile = (ROOT / "engine/dev/Dockerfile").read_text(encoding="utf-8")
    assert "./cmd/engine-mcp/" in dockerfile
    assert "./cmd/engine-rpc/" in dockerfile
    assert "COPY --from=build /out/engine-mcp /usr/local/bin/engine-mcp" in dockerfile
    assert "COPY --from=build /out/engine-rpc /usr/local/bin/engine-rpc" in dockerfile
    assert 'ENTRYPOINT ["/usr/local/bin/engine-rpc", "--addr", ":8080"]' in dockerfile

    makefile = (ROOT / "Makefile").read_text(encoding="utf-8")
    assert "build-engine-rpc:" in makefile
    assert "build: build-skill build-engine-mcp build-engine-rpc" in makefile
    assert "rm -f bin/engine-mcp bin/engine-rpc" in makefile


def test_ci_smokes_both_engine_runtimes():
    workflow = (ROOT / ".github/workflows/ci.yml").read_text(encoding="utf-8")
    assert "scripts/test_engine_mcp.py" in workflow
    assert "--entrypoint /usr/local/bin/engine-mcp" in workflow
    assert "/usr/local/bin/engine-rpc" in workflow
    assert "/jsonrpc" in workflow
