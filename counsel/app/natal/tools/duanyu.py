"""Compatibility façade for the split natal assertion domain."""

from app.domain.natal.duanyu import (
    BAZI_ONLY_RULES,
    CURRENT_LIMIT_RULES,
    NATAL_RULES,
    SCENE_ALIASES,
    SCENE_DOMAIN_FILTERS,
    YEARLY_RULES,
    ZIWEI_ONLY_RULES,
    brief,
    default_scene_domains,
    filter_domains,
    flow_factor_names,
    load_rule_table,
    load_rule_tables,
    natal_factors_for_flow,
    query,
    query_yearly,
    required_flow_factors,
    required_natal_factors,
    reset_current_year_cache,
    resolve_current_year,
    resolve_current_years,
    yearly_range,
)
from app.domain.natal.duanyu.assertion_query import current_step
from app.domain.natal.duanyu.cache import year_cache
from app.domain.natal.duanyu.yearly import _batch_liunian_data as _batch_liunian_data_impl
from app.natal.tools.duanyu_match import assert_snapshot_factors, match_rule, match_table
from app.natal.tools.factor_constants import load_constants
from app.natal.tools.yearly_eval import resolve_rules

_assert_snapshot_factors = assert_snapshot_factors
_batch_liunian_data = _batch_liunian_data_impl
_load_constants = load_constants
_resolve_rules = resolve_rules
_current_step = current_step
_year_cache = year_cache

__all__ = [
    "NATAL_RULES",
    "YEARLY_RULES",
    "SCENE_ALIASES",
    "SCENE_DOMAIN_FILTERS",
    "BAZI_ONLY_RULES",
    "ZIWEI_ONLY_RULES",
    "CURRENT_LIMIT_RULES",
    "load_rule_table",
    "load_rule_tables",
    "required_natal_factors",
    "required_flow_factors",
    "natal_factors_for_flow",
    "flow_factor_names",
    "filter_domains",
    "default_scene_domains",
    "query",
    "brief",
    "resolve_current_year",
    "resolve_current_years",
    "reset_current_year_cache",
    "resolve_rules",
    "query_yearly",
    "match_table",
    "match_rule",
    "yearly_range",
]
