"""用神 policy 的结构校验。

只校验表结构，不解释命理条件。命理规则全部位于 policy.v1.json。
"""
from __future__ import annotations

from typing import Any


class YongShenPolicyError(ValueError):
    """policy 表不满足解释器契约。"""


def validate_policy(policy: Any) -> None:
    if not isinstance(policy, dict):
        raise YongShenPolicyError("policy 必须是 object")
    for key in ("policy_id", "rule_version", "match", "tables", "defaults", "pre_factors", "rules", "finalize"):
        if key not in policy:
            raise YongShenPolicyError(f"policy 缺少字段: {key}")
    if policy["match"] != "first":
        raise YongShenPolicyError("policy.match 只支持 first")
    if not isinstance(policy["tables"], dict) or not policy["tables"]:
        raise YongShenPolicyError("policy.tables 必须是非空 object")
    if not isinstance(policy["defaults"], dict):
        raise YongShenPolicyError("policy.defaults 必须是 object")
    if not isinstance(policy["pre_factors"], list):
        raise YongShenPolicyError("policy.pre_factors 必须是 array")
    if not isinstance(policy["rules"], list) or not policy["rules"]:
        raise YongShenPolicyError("policy.rules 必须是非空 array")

    seen_derived: set[str] = set()
    for index, rule in enumerate(policy["pre_factors"]):
        if not isinstance(rule, dict):
            raise YongShenPolicyError(f"pre_factors[{index}] 必须是 object")
        rule_id = rule.get("id")
        if not isinstance(rule_id, str) or not rule_id:
            raise YongShenPolicyError(f"pre_factors[{index}] 缺少 id")
        if rule_id in seen_derived:
            raise YongShenPolicyError(f"pre_factors id 重复: {rule_id}")
        seen_derived.add(rule_id)
        if not isinstance(rule.get("when"), dict):
            raise YongShenPolicyError(f"pre_factors[{rule_id}] when 必须是 object")
        if not isinstance(rule.get("set"), dict) or not rule["set"]:
            raise YongShenPolicyError(f"pre_factors[{rule_id}] set 必须是非空 object")

    seen_rules: set[str] = set()
    previous_priority: int | None = None
    for index, rule in enumerate(policy["rules"]):
        if not isinstance(rule, dict):
            raise YongShenPolicyError(f"rules[{index}] 必须是 object")
        rule_id = rule.get("id")
        priority = rule.get("priority")
        if not isinstance(rule_id, str) or not rule_id:
            raise YongShenPolicyError(f"rules[{index}] 缺少 id")
        if rule_id in seen_rules:
            raise YongShenPolicyError(f"rules id 重复: {rule_id}")
        seen_rules.add(rule_id)
        if not isinstance(priority, int) or isinstance(priority, bool):
            raise YongShenPolicyError(f"rules[{rule_id}] priority 必须是整数")
        if previous_priority is not None and priority < previous_priority:
            raise YongShenPolicyError(f"rules[{rule_id}] priority 必须非递减")
        previous_priority = priority
        if not isinstance(rule.get("when"), dict):
            raise YongShenPolicyError(f"rules[{rule_id}] when 必须是 object")
        if not isinstance(rule.get("then"), dict) or not rule["then"]:
            raise YongShenPolicyError(f"rules[{rule_id}] then 必须是非空 object")

    finalize = policy["finalize"]
    if not isinstance(finalize, dict):
        raise YongShenPolicyError("policy.finalize 必须是 object")


__all__ = ["YongShenPolicyError", "validate_policy"]
