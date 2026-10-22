"""liki-agent-runtime runtime base 依赖 pin 契约。

liki 对 liki-agent-runtime 做依赖锁定（version + immutable digest），不做与 liki
自身版本的同名对齐；测试（runtime-validate）与本地 candidate 装配必须共用该 pin。
"""

import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PIN = ROOT / "deps" / "liki-agent-runtime.base"
SCRIPT = ROOT / "scripts" / "agent_runtime_pin.py"
RELEASE_IMAGES = ROOT / "scripts" / "release-images.sh"
MAKEFILE = ROOT / "Makefile"

CALVER_RE = re.compile(r"^\d{4}\.\d{2}\.\d{2}\.\d+$")
DIGEST_RE = re.compile(r"^sha256:[0-9a-f]{64}$")


def _pin_values() -> dict[str, str]:
    values: dict[str, str] = {}
    for line in PIN.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if line and not line.startswith("#") and ":" in line:
            key, value = line.split(":", 1)
            values[key.strip()] = value.strip()
    return values


def test_pin_file_holds_calver_and_digest():
    values = _pin_values()
    assert CALVER_RE.fullmatch(values["version"]), values
    assert DIGEST_RE.fullmatch(values["digest"]), values


def test_pin_script_check_passes():
    result = subprocess.run(
        [sys.executable, str(SCRIPT), "--check"],
        cwd=ROOT,
        capture_output=True,
        text=True,
    )
    assert result.returncode == 0, result.stderr


def test_pin_script_outputs_match_file():
    version = subprocess.run(
        [sys.executable, str(SCRIPT), "--version"],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=True,
    ).stdout.strip()
    digest = subprocess.run(
        [sys.executable, str(SCRIPT), "--digest"],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=True,
    ).stdout.strip()
    values = _pin_values()
    assert version == values["version"]
    assert digest == values["digest"]


def test_release_assembly_uses_pin_not_release_ref():
    text = RELEASE_IMAGES.read_text(encoding="utf-8")
    # 装配 base 必须来自 pin 脚本，而不是本地 mutable tag
    assert "scripts/agent_runtime_pin.py --ref" in text
    assert "BASE_IMAGE=\"$(IMAGE_REGISTRY)/liki-agent-runtime:$(IMAGE_TAG)\"" not in text


def test_makefile_agents_validate_uses_pin():
    text = MAKEFILE.read_text(encoding="utf-8")
    assert "scripts/agent_runtime_pin.py --version" in text
    assert "scripts/agent_runtime_pin.py --ref" in text
