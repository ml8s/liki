"""六爻 snapshot 投影；搬运 engine 事实并做解释层级组织，不新增命理判断。"""
from __future__ import annotations


CONCLUSION_SCOPE = "conditional_candidate_not_outcome"


def project_factors(casting: dict, chart: dict, question: dict) -> dict:
    """把 RPC raw chart 投影为稳定解释上下文。"""
    if not isinstance(casting, dict):
        raise ValueError("casting must be an object")
    if not isinstance(chart, dict):
        raise ValueError("chart must be an object")
    if not isinstance(question, dict):
        raise ValueError("question must be an object")

    yong_shen = chart.get("yong_shen")
    if not isinstance(yong_shen, dict):
        raise ValueError("chart lacks yong_shen object")

    lines = []
    for line in chart.get("lines", []):
        if not isinstance(line, dict):
            raise ValueError("chart lines contains a non-object")
        lines.append({
            "position": line.get("position"),
            "type": line.get("type"),
            "gan_zhi": f"{line.get('gan', '')}{line.get('zhi', '')}",
            "wuxing": line.get("wuxing"),
            "liu_qin": line.get("liu_qin"),
            "liu_shou": line.get("liu_shou"),
            "shi_ying": line.get("shi_ying"),
            "chang_sheng_yue": line.get("chang_sheng_yue"),
            "flags": {
                "moving": bool(line.get("dong_self")),
                "yue_po": bool(line.get("yue_po")),
                "xun_kong": bool(line.get("xun_kong")),
                "mu_ku": bool(line.get("mu_ku")),
                "mu_ku_branch": line.get("mu_ku_branch"),
                "mu_ku_element": line.get("mu_ku_element"),
                "dong_sheng": bool(line.get("dong_sheng")),
                "dong_ke": bool(line.get("dong_ke")),
            },
        })

    yong_position = yong_shen.get("position")
    yong_line = next(
        (item for item in lines if item.get("position") == yong_position),
        None,
    )
    world = next((item for item in lines if item.get("shi_ying") == "世"), None)
    other = next((item for item in lines if item.get("shi_ying") == "应"), None)

    primary = []
    secondary = []
    reference = []
    ignored_scope = []

    if yong_line:
        primary.append({
            "id": "yong-shen-state",
            "fact": {
                "position": yong_line.get("position"),
                "liu_qin": yong_line.get("liu_qin"),
                "wang_shuai": yong_shen.get("wang_shuai"),
                "yue_po": yong_shen.get("yue_po", False),
                "xun_kong": yong_shen.get("xun_kong", False),
                "mu_ku": yong_shen.get("mu_ku", False),
                "chang_sheng": yong_shen.get("chang_sheng"),
            },
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    elif yong_shen.get("is_hidden") and isinstance(yong_shen.get("fu_shen"), dict):
        primary.append({
            "id": "yong-shen-hidden-state",
            "fact": {
                "is_hidden": True,
                "position": 0,
                "fu_shen": yong_shen.get("fu_shen"),
                "wang_shuai": yong_shen.get("wang_shuai"),
                "yue_po": yong_shen.get("yue_po", False),
                "xun_kong": yong_shen.get("xun_kong", False),
                "mu_ku": yong_shen.get("mu_ku", False),
                "chang_sheng": yong_shen.get("chang_sheng"),
            },
            "conclusion_scope": CONCLUSION_SCOPE,
        })

    force_chain = chart.get("force_chain")
    if force_chain:
        primary.append({
            "id": "force-chain",
            "fact": force_chain,
            "conclusion_scope": CONCLUSION_SCOPE,
        })

    for relation in chart.get("dong_yao_relations", []):
        primary.append({
            "id": f"moving-relation-{relation.get('position')}",
            "fact": relation,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for pattern in chart.get("patterns", []):
        secondary.append({
            "id": f"pattern-{len(secondary) + 1}",
            "fact": pattern,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for fact in chart.get("day_clash_facts", []):
        secondary.append({
            "id": f"day-clash-{fact.get('position')}",
            "fact": fact,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for fact in chart.get("moving_transformations", []):
        primary.append({
            "id": f"moving-transformation-{fact.get('position')}",
            "fact": fact,
            "conclusion_scope": CONCLUSION_SCOPE,
        })

    for fact in chart.get("hidden_lines", []):
        secondary.append({
            "id": f"hidden-line-{fact.get('position')}-{fact.get('liu_qin')}",
            "fact": fact,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for fact in chart.get("branch_relation_facts", []):
        if "yong_shen" in fact.get("targets", []) or "world" in fact.get("targets", []):
            secondary.append({
                "id": f"branch-relation-{fact.get('left_position')}-{fact.get('right_position')}",
                "fact": fact,
                "conclusion_scope": CONCLUSION_SCOPE,
            })
        else:
            ignored_scope.append({
                "id": f"branch-relation-{fact.get('left_position')}-{fact.get('right_position')}",
                "reason": "未直接作用用神或世爻；仅作参考",
                "fact": fact,
            })
    for fact in chart.get("san_he_candidates", []):
        target = secondary if fact.get("targets") else ignored_scope
        target.append({
            "id": f"san-he-{fact.get('element')}-{len(target) + 1}",
            "fact": fact,
            "conclusion_scope": "structure_candidate_only_not_outcome",
        })

    reference.append({
        "id": "hexagram-name",
        "fact": {"name": chart.get("name"), "ben_gua": chart.get("ben_gua")},
        "conclusion_scope": "reference_not_primary_judgment",
    })
    for line in lines:
        reference.append({
            "id": f"liu-shou-{line.get('position')}",
            "fact": {
                "position": line.get("position"),
                "liu_shou": line.get("liu_shou"),
            },
            "conclusion_scope": "reference_not_primary_judgment",
        })
    gua_ci = chart.get("gua_ci")
    if gua_ci:
        reference.append({
            "id": "gua-ci",
            "fact": gua_ci,
            "conclusion_scope": "reference_not_primary_judgment",
        })

    conflicts = [item for item in chart.get("conflicts", []) if isinstance(item, dict)]

    evidence = {
        "primary": primary,
        "secondary": secondary,
        "reference": reference,
        "conflicts": conflicts,
        "ignored_scope": ignored_scope,
    }

    timing_candidates = chart.get("timing_candidates")
    if not isinstance(timing_candidates, list):
        timing = chart.get("ying_qi")
        timing_candidates = [timing] if isinstance(timing, dict) else []

    return {
        "schema_version": "liuyao-factors-v1",
        "question": question,
        "casting": {
            "mode": casting.get("mode"),
            "order": casting.get("order", "bottom_up"),
            "casting_id": casting.get("casting_id"),
            "rounds": casting.get("rounds", []),
            "yaos": casting.get("yaos", []),
            "dong_yao": casting.get("dong_yao", []),
        },
        "board": {
            "name": chart.get("name"),
            "ben_gua": chart.get("ben_gua"),
            "bian_gua": chart.get("bian_gua"),
            "palace": chart.get("gong"),
            "palace_wuxing": chart.get("gong_wuxing"),
            "lines": lines,
        },
        "focus": {
            "yong_shen": yong_shen,
            "yong_line": yong_line,
            "world": world,
            "other": other,
        },
        "evidence": evidence,
        "facts": {
            "hidden_lines": chart.get("hidden_lines", []),
            "branch_relation_facts": chart.get("branch_relation_facts", []),
            "day_clash_facts": chart.get("day_clash_facts", []),
            "moving_transformations": chart.get("moving_transformations", []),
            "san_he_candidates": chart.get("san_he_candidates", []),
            "force_chain": chart.get("force_chain"),
            "yong_shen_candidates": chart.get("yong_shen_candidates", []),
        },
        "timing_candidates": timing_candidates,
        "policy": {
            "may_output": ["tendency", "condition", "timing_window", "action"],
            "must_cover_conflicts": True,
            "forbidden": ["absolute_guarantee"],
        },

    }
