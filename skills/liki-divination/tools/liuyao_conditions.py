"""六爻歌诀条件还原；把流行断语还原为可核对条件，不输出绝对结论。"""
from __future__ import annotations

import json
from functools import lru_cache
from pathlib import Path


PATH = Path(__file__).with_name("liuyao_condition_rules.json")


@lru_cache
def load_rules() -> dict:
    return json.loads(PATH.read_text(encoding="utf-8"))


def _get_path(source: dict, path: str):
    current = source
    for part in path.split("."):
        if not isinstance(current, dict) or part not in current:
            return None
        current = current[part]
    return current


def _rule_state(rule: dict, snapshot: dict, topic: str | None) -> tuple[bool, str, list[str]]:
    applies = rule.get("applies_if", {})
    for path, expected in applies.items():
        if path == "topic":
            if topic != expected:
                return False, "not_applicable", []
        elif path == "moving_relation.relation":
            continue
        else:
            if _get_path(snapshot, path) != expected:
                return False, "not_applicable", []
    yong_line = snapshot.get("focus", {}).get("yong_line", {}) or {}
    state_rules = load_rules()["state_classes"]
    strong = _matches_state_class(snapshot, state_rules["strong"])
    weak = _matches_state_class(snapshot, state_rules["weak"])
    moving = bool((yong_line.get("flags") or {}).get("moving"))
    missing = []
    if strong:
        branch = rule.get("strong_branch", {})
        return True, "strong", branch.get("requires", missing)
    if weak:
        branch = rule.get("weak_branch", {})
        return True, "weak", branch.get("requires", missing)
    if path_has_moving_relation(snapshot, rule):
        return True, "conditional", ["须核对作用爻旺衰与是否实际作用用神"]
    return True, "conditional", list(missing)


def _matches_state_class(snapshot: dict, rule: dict) -> bool:
    return _get_path(snapshot, rule["path"]) in set(rule["values"])


def path_has_moving_relation(snapshot: dict, rule: dict) -> bool:
    applies = rule.get("applies_if", {})
    expected = applies.get("moving_relation.relation")
    if not expected:
        return True
    for item in snapshot.get("evidence", {}).get("primary", []):
        if not isinstance(item.get("fact"), dict):
            continue
        fact = item.get("fact", {})
        if fact.get("relation") == expected:
            return True
    return False


def evaluate(snapshot: dict, topic: str | None = None) -> dict:
    if not isinstance(snapshot, dict):
        raise ValueError("snapshot must be an object")
    results = []
    for rule in load_rules().get("rules", []):
        applies, state, requires = _rule_state(rule, snapshot, topic)
        branch = {}
        if state == "strong":
            branch = rule.get("strong_branch", {})
        elif state == "weak":
            branch = rule.get("weak_branch", {})
        elif applies:
            branch = {"conclusion": rule.get("restored", "")}
        results.append({
            "id": rule["id"],
            "statement": rule["statement"],
            "restored": rule["restored"],
            "applies": applies,
            "state": state,
            "conclusion": branch.get("conclusion", "") if applies else "",
            "requires": requires,
            "forbidden": rule.get("forbidden", []),
            "conclusion_scope": "conditional_rule_restoration_not_absolute_verdict",
        })
    return {
        "schema_version": load_rules()["schema_version"],
        "topic": topic,
        "results": results,
        "policy": {
            "must_restore_hidden_conditions": True,
            "forbid_absolute_old_phrases": True,
            "reference_may_not_override_primary": True,
        },
    }
