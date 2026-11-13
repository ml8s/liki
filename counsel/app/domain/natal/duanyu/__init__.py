"""Canonical natal assertion query domain."""

from .assertion_query import (
    brief,
    default_scene_domains,
    filter_domains,
    query,
    query_yearly,
)
from .cache import reset_current_year_cache, resolve_current_year, resolve_current_years
from .factor_closure import (
    flow_factor_names,
    natal_factors_for_flow,
    required_flow_factors,
    required_natal_factors,
)
from .rule_loader import (
    BAZI_ONLY_RULES,
    CURRENT_LIMIT_RULES,
    NATAL_RULES,
    SCENE_ALIASES,
    SCENE_DOMAIN_FILTERS,
    YEARLY_RULES,
    ZIWEI_ONLY_RULES,
    load_rule_table,
    load_rule_tables,
)
from .yearly import yearly_range

__all__ = [
    "BAZI_ONLY_RULES",
    "CURRENT_LIMIT_RULES",
    "NATAL_RULES",
    "SCENE_ALIASES",
    "SCENE_DOMAIN_FILTERS",
    "YEARLY_RULES",
    "ZIWEI_ONLY_RULES",
    "brief",
    "default_scene_domains",
    "filter_domains",
    "flow_factor_names",
    "load_rule_table",
    "load_rule_tables",
    "natal_factors_for_flow",
    "query",
    "query_yearly",
    "required_flow_factors",
    "required_natal_factors",
    "reset_current_year_cache",
    "resolve_current_year",
    "resolve_current_years",
    "yearly_range",
]
