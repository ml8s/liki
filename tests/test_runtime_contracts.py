
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


def test_engine_image_carries_mcp_runtime_only():
    dockerfile = (ROOT / "engine/dev/Dockerfile").read_text(encoding="utf-8")
    assert "./cmd/engine-mcp/" in dockerfile
    assert "COPY --from=build /out/engine-mcp /usr/local/bin/engine-mcp" in dockerfile
    assert 'ENTRYPOINT ["/usr/local/bin/engine-mcp", "--addr", ":8081"]' in dockerfile
    assert "engine-rpc" not in dockerfile

    makefile = (ROOT / "Makefile").read_text(encoding="utf-8")
    assert "build-engine-rpc:" not in makefile
    assert "build: build-skill build-engine-mcp" in makefile
    assert "rm -f bin/engine-mcp" in makefile


def test_release_smokes_engine_mcp_runtime():
    smoke = (ROOT / "scripts/smoke-image.sh").read_text(encoding="utf-8")
    assert "scripts/test_engine_mcp.py" in smoke
    assert "--entrypoint /usr/local/bin/engine-mcp" in smoke
    assert "engine-rpc" not in smoke
    assert "/jsonrpc" not in smoke
