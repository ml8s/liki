"""Current divination case/projection guardrails shared across toolfaces."""

import pytest
from app.contracts.divination import DivinationConflict
from app.divination.tools.divination_safety import assess
from app.divination.tools.liuyao_factors import project_factors
from pydantic import ValidationError


def test_short_high_risk_question_is_not_blocked_by_safety():
    safety = assess("我不想活")
    assert safety["status"] == "advisory"
    assert safety["support_first"] is True


def test_short_high_risk_message_is_not_blocked_by_safety():
    safety = assess("自杀")
    assert safety["status"] == "advisory"
    assert safety["support_first"] is True


def test_conflict_contract_accepts_facts_omitted_and_rejects_empty():
    conflict = DivinationConflict.model_validate({
        "id": "strong-but-void", "reason": "须辨真假空",
    })
    assert conflict.facts == []

    with pytest.raises(ValidationError):
        DivinationConflict.model_validate({"id": "", "reason": "x"})
    with pytest.raises(ValidationError):
        DivinationConflict.model_validate({
            "id": "x", "reason": "y", "facts": [""],
        })
    with pytest.raises(ValidationError):
        DivinationConflict.model_validate({
            "id": "x", "reason": "y", "unknown": True,
        })


def test_liuyao_factors_fail_closed_on_malformed_engine_conflicts():
    def project(conflicts):
        return project_factors(
            casting={},
            chart={"yong_shen": {"name": "官鬼", "position": 4}, "conflicts": conflicts},
            question={"text": "测事"},
        )

    valid = project([{"id": "strong-but-void", "reason": "须辨真假空"}])
    assert valid["evidence"]["conflicts"] == [{
        "id": "strong-but-void", "reason": "须辨真假空", "facts": [],
    }]
    for malformed in ("not-a-list", [{"reason": "missing id"}], [1]):
        with pytest.raises(ValueError, match="liuyao chart conflicts are invalid"):
            project(malformed)
