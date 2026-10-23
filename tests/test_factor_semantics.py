"""因子语义内核的封闭四态契约。"""

import pytest
from app.natal.tools.errors import FactorEvaluateError
from app.natal.tools.factor_semantics import PredicateState, all_of, any_of, factor_value, predicate

STATES = list(PredicateState)


def test_factor_value_normalizes_legacy_wire_values() -> None:
    assert factor_value(1) == PredicateState.HIT
    assert factor_value(0) == PredicateState.MISS
    assert factor_value("unknown") == PredicateState.UNKNOWN
    assert factor_value("") == PredicateState.UNKNOWN
    assert factor_value("印星") == "印星"


@pytest.mark.parametrize("value", [True, False, 2, None, [], {}, object()])
def test_factor_value_rejects_non_domain_shapes(value: object) -> None:
    with pytest.raises(FactorEvaluateError, match="FactorEvaluateError|不能返回|不是 FactorValue"):
        factor_value(value, where="test")


def test_predicate_preserves_non_applicable_and_unknown() -> None:
    assert predicate(PredicateState.NOT_APPLICABLE, "印星") == PredicateState.NOT_APPLICABLE
    assert predicate(PredicateState.UNKNOWN, "印星") == PredicateState.UNKNOWN
    assert predicate("印星", "印星") == PredicateState.HIT
    assert predicate("财星", "印星") == PredicateState.MISS


def test_all_of_uses_closed_four_state_algebra() -> None:
    for left in STATES:
        for right in STATES:
            expected = PredicateState.MISS
            if PredicateState.MISS not in (left, right):
                if PredicateState.UNKNOWN in (left, right):
                    expected = PredicateState.UNKNOWN
                elif PredicateState.NOT_APPLICABLE in (left, right):
                    expected = PredicateState.NOT_APPLICABLE
                else:
                    expected = PredicateState.HIT
            assert all_of([left, right]) == expected


def test_any_of_uses_closed_four_state_algebra() -> None:
    for left in STATES:
        for right in STATES:
            expected = PredicateState.NOT_APPLICABLE
            if PredicateState.HIT in (left, right):
                expected = PredicateState.HIT
            elif PredicateState.MISS in (left, right):
                expected = PredicateState.MISS
            elif PredicateState.UNKNOWN in (left, right):
                expected = PredicateState.UNKNOWN
            assert any_of([left, right]) == expected
