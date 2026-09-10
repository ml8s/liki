"""Guard the boundary between the accuracy benchmark and behavior smoke suite."""

import importlib.util
from pathlib import Path

import yaml


ROOT = Path(__file__).resolve().parents[1]
BENCHMARK = ROOT / "tests/benchmark/mingli160"
SMOKE = ROOT / "tests/skillup"
DOMAINS = ("bazi", "divination", "fengshui", "naming")


def _load_grade_rules():
    path = SMOKE / "grade.py"
    spec = importlib.util.spec_from_file_location("skillup_grade", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module.RULES


def test_accuracy_benchmark_is_isolated_from_behavior_smoke() -> None:
    assert (BENCHMARK / "evals/eval.yaml").is_file()
    assert (BENCHMARK / "grade-case.py").is_file()
    assert (BENCHMARK / "answers.json").is_file()
    assert not (SMOKE / "evals/eval.yaml").exists()
    assert not (SMOKE / "answers.json").exists()


def test_skillup_smoke_cases_are_complete_and_contract_guarded() -> None:
    rules = _load_grade_rules()
    seen_ids: set[str] = set()
    for domain in DOMAINS:
        config = yaml.safe_load((SMOKE / f"evals/{domain}.yaml").read_text(encoding="utf-8"))
        assert config["schema_version"] == "v1alpha1"
        assert config["environment"]["type"] == "docker"
        assert config["engine"]["name"] == "qwen_code"

        case_files = config["cases"]["files"]
        assert case_files, f"{domain} smoke suite is empty"
        for relative in case_files:
            case_path = SMOKE / relative
            case = yaml.safe_load(case_path.read_text(encoding="utf-8"))
            case_id = case["id"]
            assert case_id not in seen_ids, case_id
            seen_ids.add(case_id)
            assert case_id in rules, f"missing behavior rule: {case_id}"
            assert f"SMOKE_ID: {case_id}" in case["input"]["prompt"]
            assert case["judge"]["type"] == "script"
            assert case["judge"]["script_path"] == "grade.py"

    assert len(seen_ids) == 21
    assert set(rules) == seen_ids
