from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"


def test_deprecated_rpc_wrapper_is_gone():
    assert not (TOOLS / "liuyao_rpc.py").exists()
    for path in TOOLS.glob("*.py"):
        assert "liuyao_rpc" not in path.read_text(encoding="utf-8")


def test_only_common_rpc_module_touches_urllib():
    rpc_files = []
    for path in TOOLS.glob("*.py"):
        text = path.read_text(encoding="utf-8")
        if "urllib.request" in text or "urlopen(" in text:
            rpc_files.append(path.name)
    assert rpc_files == ["divination_rpc.py"]


def test_primary_entries_use_shared_safety():
    for filename in (
        "divination_router.py",
        "liuyao_read.py",
        "qimen_read.py",
        "huangli_days.py",
    ):
        text = (TOOLS / filename).read_text(encoding="utf-8")
        assert "divination_safety" in text, f"{filename} must use shared safety"
