"""一次性奇门问测：创建内部 case 并返回受控 Reading。"""

from app.divination.tools.qimen_answer_core import build as build_standard_core
from app.divination.tools.qimen_answer_core import validate_core as validate_standard_core
from app.divination.tools.qimen_case import create as create_case
from app.divination.tools.qimen_jinhan_answer_core import build as build_jinhan_core
from app.divination.tools.qimen_jinhan_answer_core import validate_core as validate_jinhan_core
from app.divination.tools.qimen_jinhan_factors import validate as validate_jinhan_factors
from app.divination.tools.qimen_projection import validate as validate_standard_factors
from app.divination.tools.qimen_projection_common import public_palace_keys
from app.guards import require_question

_ALL_QIMEN_DOMAINS = {
    "career", "health", "hiding", "legal", "missing_person",
    "relationship", "study", "travel", "wealth", "lost_property",
    "thief_capture", "capture_escape", "thief_profile", "timing",
    "palace_relation",
}

_FOLLOWUP_POLICY = {
    "same_matter": {
        "clarification": "allowed_from_reading",
        "timing": "allowed_from_reading",
        "obstacle": "allowed_from_reading",
    },
    "new_domain": "requires_new_case",
    "another_matter": "requires_new_case",
    "recast_same_matter": "forbidden",
    "beyond_returned_evidence": "answer_insufficient",
}


def _build_scope(case: dict) -> dict:
    focus = case.get("focus")
    special = case.get("special")
    focus_kind = focus.get("kind") if isinstance(focus, dict) else None
    focus_value = focus.get("matter") if focus_kind == "matter" else (
        ",".join(focus.get("yong_shen", []))
        if focus_kind == "yong_shen" else None
    )
    rule = special.get("rule") if isinstance(special, dict) else None
    included = [value for value in (focus_value, rule, "timing", "palace_relation") if value]
    return {
        "focus_kind": focus_kind,
        "focus_value": focus_value,
        "rule": rule,
        "included_domains": included,
        "excluded_domains": sorted(_ALL_QIMEN_DOMAINS - set(included)),
        "followup_policy": _FOLLOWUP_POLICY,
    }


def _fact_by_ref(case: dict, fact_ref: str):
    if not fact_ref.startswith("factors."):
        return None
    key = fact_ref.removeprefix("factors.")
    raw_key = key
    if key == "palace_wang_shuai":
        raw_key = "gong_wang_shuai"
    elif "_palace_" in key and key not in {"day_stem_palace", "hour_stem_palace"}:
        raw_key = key.replace("_palace_", "_gong_", 1)
    elif key.startswith("palace_"):
        raw_key = "gong_" + key[len("palace_"):]
    elif key.endswith("_palace") and key not in {"day_stem_palace", "hour_stem_palace"}:
        raw_key = key[:-len("_palace")] + "_gong"
    return public_palace_keys(case.get("factors", {}).get(raw_key))


def _public_fact_ref(fact_ref: str) -> str:
    if not fact_ref.startswith("factors."):
        return fact_ref
    key = fact_ref.removeprefix("factors.")
    if key.startswith("gong_"):
        key = "palace_" + key[len("gong_"):]
    elif key.endswith("_gong"):
        key = key[:-len("_gong")] + "_palace"
    return f"factors.{key}"


def _compact_relation_fact(source: str, fact: object) -> object:
    """Group one-to-many palace relations without dropping any pair."""
    relation_kind = source.rsplit(".", 1)[-1]
    if relation_kind not in {
        "spirit_star_relations", "spirit_spirit_relations", "spirit_door_relations",
    } or not isinstance(fact, list):
        return fact

    target_key = {
        "spirit_star_relations": "star",
        "spirit_spirit_relations": "other",
        "spirit_door_relations": "door",
    }[relation_kind]
    left_fields = {"spirit", "spirit_name", "spirit_palace"}
    left_order: list[tuple[object, ...]] = []
    left_by_key: dict[tuple[object, ...], dict] = {}

    for row in fact:
        if not isinstance(row, dict):
            return fact
        try:
            left_key = (
                row["spirit"], row.get("spirit_name"),
                row["spirit_palace"],
            )
        except KeyError:
            return fact
        if left_key not in left_by_key:
            left_by_key[left_key] = {
                "spirit": row["spirit"],
                "spirit_name": row.get("spirit_name"),
                "spirit_palace": row["spirit_palace"],
                "relations": {},
            }
            left_order.append(left_key)
        left = left_by_key[left_key]
        target = {
            key: value for key, value in row.items()
            if key not in left_fields | {"relation", "same_palace"}
            and key != target_key
            and not key.endswith("_palace")
        }
        target[target_key] = row[target_key]
        target["palace"] = row[f"{target_key}_palace"]
        target["same_palace"] = row["same_palace"]
        left["relations"].setdefault(row["relation"], []).append(target)

    return [left_by_key[key] for key in left_order]


def _assertion_by_ref(case: dict, ref: str):
    if not ref.startswith("assertion:"):
        return None
    assertion_id = ref.removeprefix("assertion:")
    special = case.get("special") or {}
    for item in special.get("assertions", []):
        if item.get("id") == assertion_id:
            return {"source": "special", **item}
    for item in (case.get("judgments") or {}).get("results", []):
        if item.get("id") == assertion_id:
            return {"source": "judgment", **item}
    return None


def _project_evidence(case: dict, answer: dict) -> list[dict]:
    projected = []
    for ref in answer.get("refs", []):
        if ref.startswith("case:"):
            evidence_id = ref
            item = next(
                (entry for entry in case.get("evidence", []) if entry.get("id") == ref),
                None,
            )
            if item is None:
                continue
            fact = item.get("fact") or _fact_by_ref(case, item.get("fact_ref", ""))
            source = _public_fact_ref(item.get("fact_ref", ""))
            fact = _compact_relation_fact(source, public_palace_keys(fact))
            projected.append({
                "id": evidence_id,
                "fact_ref": source,
                "layer": item.get("layer", "context"),
                "fact": fact,
            })
            continue

        assertion = _assertion_by_ref(case, ref)
        if assertion is None:
            continue
        projected.append({
            "id": ref,
            "layer": "assertion",
            "fact": assertion,
        })
    return projected


def _reading(case: dict, answer: dict) -> dict:
    reading = {
        "technique": "qimen",
        "question": case["question"],
        "scope": _build_scope(case),
        "qi_ju": case["qi_ju"],
        "focus": case.get("focus"),
        "method": case["method"],
        "evidence": _project_evidence(case, answer),
        "timing_candidates": [
            candidate
            for candidate in case.get("timing_candidates", [])
        ],
        "special_assertions": (case.get("special") or {}).get("assertions", []),
        "judgment_assertions": [
            {
                key: value for key, value in item.items()
                if key != "applies"
            }
            for item in (case.get("judgments") or {}).get("results", [])
            if item.get("applies") is True
        ],
        "interpretation": {
            "headline": answer["headline"],
            "verdict": answer["verdict"],
            "action": answer["action"],
            "boundary": answer["boundary"],
            "disclaimer": answer["disclaimer"],
        },
        "safety_advisory": answer["safety_advisory"],
    }
    return reading


def _answer(case: dict) -> dict:
    method = case["method"]
    factors = {
        "scope": method["scope"],
        "school": method["school"],
        "ying_qi": [],
        **case["factors"],
    }
    if factors.get("school") == "jinhan_yujing":
        validate_jinhan_factors(factors)
        core = build_jinhan_core(case)
    else:
        validate_standard_factors(factors)
        core = build_standard_core(case)

    timing_count = len(core.get("timing_refs", []))
    core.update({
        "headline": "奇门局例条件性解读",
        "verdict": (
            f"当前局提供 {len(core.get('refs', []))} 项盘面/断语因子"
            f"和 {timing_count} 个应期候选；只能表达态势倾向，不构成结果承诺。"
        ),
        "action": "先核查现实约束，再结合候选方向或时机推进。",
    })
    audit = (
        validate_jinhan_core(core, case)
        if factors.get("school") == "jinhan_yujing"
        else validate_standard_core(core, case)
    )
    if not audit.get("accepted"):
        errors = "; ".join(
            str(item.get("reason") or item.get("field") or item)
            for item in audit.get("errors", [])
        )
        raise ValueError(f"qimen answer failed validation: {errors}")
    return {
        "safety_advisory": case["safety_advisory"],
        "headline": core["headline"],
        "verdict": core["verdict"],
        "refs": core["refs"],
        "timing_refs": core["timing_refs"],
        "action": core["action"],
        "boundary": core["boundary"],
        "disclaimer": core["disclaimer"],
    }


def ask(**arguments) -> dict:
    """创建一次奇门内部 case，并返回自包含受控 Reading。"""
    require_question(arguments["question"])
    case = create_case(**arguments)
    answer = _answer(case)
    return _reading(case, answer)
