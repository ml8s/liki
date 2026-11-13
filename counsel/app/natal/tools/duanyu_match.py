"""Compatibility façade for the natal assertion matching domain."""

from app.domain.natal.duanyu.matching import (
    assert_snapshot_factors,
    match_rule,
    match_table,
)

__all__ = [
    "match_table",
    "match_rule",
    "assert_snapshot_factors",
]
