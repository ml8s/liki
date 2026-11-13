"""Rule-table and domain configuration loading for natal assertions."""

from app.domain.natal.models import Rule
from app.natal.tools.assertion_store import load_rule_table
from app.natal.tools.factor_constants import load_constants
from app.natal.tools.yearly_eval import resolve_rules

DOMAIN_CONFIG = load_constants()["命理域"]
NATAL_RULES = frozenset(DOMAIN_CONFIG["本命规则"])
YEARLY_RULES = frozenset(DOMAIN_CONFIG["流年规则"])
SCENE_ALIASES = {alias: tuple(rules) for alias, rules in DOMAIN_CONFIG["场景别名"].items()}
SCENE_DOMAIN_FILTERS = {alias: tuple(domains) for alias, domains in DOMAIN_CONFIG.get("场景领域过滤", {}).items()}
BAZI_ONLY_RULES = frozenset(DOMAIN_CONFIG["八字专属域"])
ZIWEI_ONLY_RULES = frozenset(DOMAIN_CONFIG["紫微专属域"])
CURRENT_LIMIT_RULES = frozenset(DOMAIN_CONFIG["当前限运规则"])


_RULE_MODEL_CACHE: dict[tuple[str, str], tuple[Rule, ...]] = {}


def load_rule_models(rule: str) -> dict[str, tuple[Rule, ...]]:
    """Load immutable rule models once per process and side."""
    return {side: _rule_models(rule, side, rows) for side, rows in load_rule_tables(rule).items()}


def _rule_models(rule: str, side: str, rows: list[dict]) -> tuple[Rule, ...]:
    key = (rule, side)
    if key not in _RULE_MODEL_CACHE:
        _RULE_MODEL_CACHE[key] = tuple(Rule.from_row(row) for row in rows)
    return _RULE_MODEL_CACHE[key]


def load_rule_tables(rule: str) -> dict[str, list[dict]]:
    """按命理侧配置加载一个域的三侧断言表。"""
    side_config = load_constants()["命理侧"]
    tables = {}
    for side in side_config["断言代码"]:
        if side == "bazi":
            required = rule not in ZIWEI_ONLY_RULES
        else:
            required = rule not in BAZI_ONLY_RULES
        tables[side] = load_rule_table(f"{side}_{rule}.csv", required=required)
    return tables


def assertion_domains(tables: list[list[dict]]) -> set[str]:
    """读取规则表自带的领域闭集；不硬编码领域清单。"""
    return {row["领域"] for table in tables for row in table if row.get("领域")}


def resolve_rule_selection(rules):
    return resolve_rules(rules, YEARLY_RULES, SCENE_ALIASES)
