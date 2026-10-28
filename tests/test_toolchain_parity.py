"""Toolchain pins must be single-sourced: local `make ...` and CI run the same 版本.

一个工具两个版本 = 本地过、CI 挂。这里把每个 pin 与其消费者（本地脚本、CI job）
钉在一起，任一处漂移都会失败。覆盖 golangci-lint、ruff 和 engine 覆盖率门槛。
"""

import re
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
WORKFLOW = ROOT / ".github" / "workflows" / "ci.yml"
MAKEFILE = ROOT / "Makefile"
GOLANGCI_PIN = ROOT / "deps" / "golangci-lint.version"
LINT_ENGINE = ROOT / "scripts" / "lint-engine.sh"
LINT_PYTHON = ROOT / "scripts" / "lint-python.sh"
COUNSEL_TEST = ROOT / "scripts" / "test-counsel.sh"
ENGINE_COVERAGE = ROOT / "scripts" / "test-engine-coverage.sh"
PRE_PUSH_HOOK = ROOT / ".githooks" / "pre-push"


def _workflow() -> dict:
    return yaml.safe_load(WORKFLOW.read_text(encoding="utf-8"))


def _steps_using(action: str) -> list[dict]:
    steps = []
    for job in (_workflow().get("jobs") or {}).values():
        for step in job.get("steps") or []:
            if action in str(step.get("uses", "")):
                steps.append(step)
    return steps


def test_golangci_pin_is_calver_with_leading_v():
    value = re.search(r"^version:\s*(\S+)\s*$", GOLANGCI_PIN.read_text(encoding="utf-8"), re.M)
    assert value, "deps/golangci-lint.version must carry a `version:` line"
    assert re.fullmatch(r"v\d+\.\d+\.\d+", value.group(1)), value.group(1)


def test_ci_golangci_action_matches_pin():
    pinned = re.search(
        r"^version:\s*(\S+)\s*$", GOLANGCI_PIN.read_text(encoding="utf-8"), re.M
    ).group(1)
    steps = _steps_using("golangci/golangci-lint-action")
    assert steps, "CI must run golangci-lint"
    versions = {str(step.get("with", {}).get("version")) for step in steps}
    assert versions == {pinned}, versions


def test_makefile_lint_engine_uses_the_pinned_script():
    makefile = MAKEFILE.read_text(encoding="utf-8")
    target = re.search(r"^lint-engine:.*\n((?:\t.*\n)+)", makefile, re.M)
    assert target, "lint-engine target missing"
    body = target.group(1)
    assert "scripts/lint-engine.sh" in body, body
    assert "golangci-lint run" not in body, "engine lint must go through the pin assertion"


def test_lint_engine_script_asserts_the_pin():
    script = LINT_ENGINE.read_text(encoding="utf-8")
    assert "deps/golangci-lint.version" in script
    assert "install.sh | sh -s -- -b" in script, "must say how to fix a mismatch"
    assert "go install github.com" not in script, "CONTRIBUTING.md forbids go install for golangci-lint"


def test_ruff_pin_matches_ci():
    script = LINT_PYTHON.read_text(encoding="utf-8")
    local = re.search(r'RUFF_VERSION="\$\{RUFF_VERSION:-([^}"]+)\}"', script)
    assert local, "lint-python.sh must default RUFF_VERSION"
    ci = re.search(r"ruff==([0-9][0-9.]*)", WORKFLOW.read_text(encoding="utf-8"))
    assert ci, "CI must install a pinned ruff"
    assert local.group(1) == ci.group(1), (local.group(1), ci.group(1))


def test_lint_python_script_asserts_global_ruff_pin():
    script = LINT_PYTHON.read_text(encoding="utf-8")
    assert "ruff --version" in script, "a global Ruff must be version-checked"
    assert "does not match the pinned" in script, "a version mismatch must fail closed"
    assert "exec ruff check" in script


def test_counsel_bootstrap_installs_dev_dependencies_on_every_toolchain_path():
    script = COUNSEL_TEST.read_text(encoding="utf-8")
    makefile = MAKEFILE.read_text(encoding="utf-8")
    assert "requirements-dev.txt" in script, (
        "the counsel suite needs pytest/coverage from requirements-dev.txt"
    )
    assert "make --no-print-directory -C \"$ROOT\" python-env" in script
    assert "uv pip install --python counsel/.venv/bin/python -r counsel/requirements-dev.txt" in makefile, (
        "the uv-provisioned venv must include dev dependencies"
    )
    assert "python" in script and "-m pip install" in script, (
        "the fallback path must install dev dependencies"
    )


def test_engine_coverage_floor_is_single_sourced():
    script = ENGINE_COVERAGE.read_text(encoding="utf-8")
    floor = re.search(r'FLOOR="\$\{COVERAGE_FLOOR:-([0-9]+)\}"', script)
    assert floor, "test-engine-coverage.sh must default COVERAGE_FLOOR"
    workflow = WORKFLOW.read_text(encoding="utf-8")
    assert "make test-engine-coverage" in workflow, "CI must run the shared coverage target"
    assert re.search(r"m=\"85\"", workflow) is None, (
        "the coverage floor must live in test-engine-coverage.sh, not CI"
    )


def test_makefile_test_engine_coverage_uses_the_script():
    makefile = MAKEFILE.read_text(encoding="utf-8")
    target = re.search(r"^test-engine-coverage:.*\n((?:\t.*\n)+)", makefile, re.M)
    assert target, "test-engine-coverage target missing"
    body = target.group(1)
    assert "scripts/test-engine-coverage.sh" in body, body
    assert "go test" not in body, "the coverage run must go through the script"


def test_pre_push_hook_runs_gate():
    """pre-push 钩子 = 显式 make gate：推上去 CI 就是绿的。"""
    assert PRE_PUSH_HOOK.is_file(), "restore .githooks/pre-push (make hooks)"
    assert "make gate" in PRE_PUSH_HOOK.read_text(encoding="utf-8")


def test_ci_provisions_go_from_engine_go_mod():
    """Go 版本唯一真相源 = engine/go.mod。"""
    text = (ROOT / ".github" / "workflows" / "ci.yml").read_text(encoding="utf-8")
    assert "go-version-file: engine/go.mod" in text
