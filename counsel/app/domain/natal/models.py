"""Immutable domain models for natal factors and assertion rules."""

from __future__ import annotations

from collections.abc import Mapping
from dataclasses import dataclass
from types import MappingProxyType
from typing import Any

from app.natal.tools.factor_semantics import FactorValue

__all__ = [
    "AssertionHit",
    "AssertionRule",
    "Condition",
    "ConditionGroup",
    "FactorSnapshot",
]


@dataclass(frozen=True, slots=True)
class Condition:
    """One factor assertion. A missing factor fails closed before matching."""

    factor: str
    expected: FactorValue


@dataclass(frozen=True, slots=True)
class ConditionGroup:
    """条件组内 AND，条件组之间 OR。"""

    conditions: tuple[Condition, ...]

    @classmethod
    def from_mapping(cls, conditions: Mapping[str, FactorValue]) -> ConditionGroup:
        return cls(tuple(Condition(factor, expected) for factor, expected in conditions.items()))

    def to_mapping(self) -> dict[str, FactorValue]:
        return {condition.factor: condition.expected for condition in self.conditions}

    def matches(self, factors: Mapping[str, FactorValue]) -> bool:
        return all(
            condition.factor in factors and factors[condition.factor] == condition.expected
            for condition in self.conditions
        )

    def trace(self, factors: Mapping[str, FactorValue], group_number: int) -> dict[str, Any]:
        return {
            "condition_group": group_number,
            "factors": {
                condition.factor: {"expected": condition.expected, "actual": factors.get(condition.factor)}
                for condition in self.conditions
            },
        }


@dataclass(frozen=True, slots=True)
class AssertionRule:
    """A canonical assertion row and its OR/AND condition groups."""

    id: str
    domain: str
    event_type: str
    time_layer: str
    event: str
    conclusion: str
    evidence: str
    classic_evidence: str
    condition_groups: tuple[ConditionGroup, ...]

    @classmethod
    def from_row(cls, row: Mapping[str, Any]) -> AssertionRule:
        groups = tuple(
            ConditionGroup.from_mapping(conditions)
            for conditions in (row.get("约束组") or [])
            if isinstance(conditions, Mapping)
        )
        return cls(
            id=str(row.get("id", "")),
            domain=str(row.get("领域", "")),
            event_type=str(row.get("事件类型", "")),
            time_layer=str(row.get("时间层", "")),
            event=str(row.get("事件", "")),
            conclusion=str(row.get("结论", "")),
            evidence=str(row.get("依据", "")),
            classic_evidence=str(row.get("经典依据", "")),
            condition_groups=groups,
        )

    def to_row(self) -> dict[str, Any]:
        """Project to the internal wire shape consumed by judgments today."""
        return {
            "id": self.id,
            "领域": self.domain,
            "事件类型": self.event_type,
            "时间层": self.time_layer,
            "事件": self.event,
            "约束组": [group.to_mapping() for group in self.condition_groups],
            "结论": self.conclusion,
            "依据": self.evidence,
            "经典依据": self.classic_evidence,
        }


@dataclass(frozen=True, slots=True)
class AssertionHit:
    """A matched assertion with explicit condition provenance."""

    rule: AssertionRule
    matched: tuple[tuple[ConditionGroup, Mapping[str, FactorValue]], ...]

    def to_row(self) -> dict[str, Any]:
        row = self.rule.to_row()
        row["trace"] = [
            group.trace(factors, group_number) for group_number, (group, factors) in enumerate(self.matched, 1)
        ]
        return row


@dataclass(frozen=True, slots=True)
class FactorSnapshot:
    """Published factors plus non-factor matching context for one assertion side."""

    factors: Mapping[str, FactorValue]
    context: Mapping[str, Any]

    def __post_init__(self) -> None:
        # Callers may pass mutable dicts. Keep the domain snapshot read-only while
        # preserving JSON-compatible values and avoiding recursive copies.
        object.__setattr__(self, "factors", MappingProxyType(dict(self.factors)))
        object.__setattr__(self, "context", MappingProxyType(dict(self.context)))
