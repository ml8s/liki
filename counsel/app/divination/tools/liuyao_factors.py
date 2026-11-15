"""六爻 case 因子投影；搬运 engine 事实并做解释层级组织，不新增命理判断。"""

from pydantic import TypeAdapter, ValidationError

from app.contracts.divination import DivinationConflict

CONCLUSION_SCOPE = "conditional_candidate_not_outcome"


def _with_id(item: dict, fact_id: str) -> dict:
    return {"id": fact_id, **item}


def project_factors(
    *,
    casting: dict,
    chart: dict,
    analysis: dict,
    question: dict,
) -> dict:
    """把 engine chart/analysis 投影为稳定解释上下文。"""
    if not isinstance(casting, dict):
        raise ValueError("casting must be an object")
    if not isinstance(chart, dict):
        raise ValueError("chart must be an object")
    if not isinstance(analysis, dict):
        raise ValueError("analysis must be an object")
    if not isinstance(question, dict):
        raise ValueError("question must be an object")

    yong_shen = chart.get("yong_shen")
    if not isinstance(yong_shen, dict):
        raise ValueError("chart lacks yong_shen object")

    def project_line(line: dict) -> dict:
        return {
            "position": line.get("position"),
            "yao_value": line.get("yao_value"),
            "gan_zhi": line.get("gan_zhi", ""),
            "wuxing": line.get("wuxing"),
            "liu_qin": line.get("liu_qin"),
            "liu_shou": line.get("liu_shou"),
            "shi_ying": line.get("shi_ying", ""),
            "chang_sheng_yue": line.get("chang_sheng_yue"),
            "wang_shuai": line.get("wang_shuai"),
            "ri_chen_relations": line.get("ri_chen_relations"),
            "flags": dict(line.get("flags", {})),
        }

    lines = [project_line(line) for line in chart.get("lines", [])]
    yong_position = yong_shen.get("position")
    yong_line = next(
        (item for item in lines if item.get("position") == yong_position),
        None,
    )
    primary = []
    secondary = []
    reference = []
    reference_only_facts = []

    if yong_line:
        primary.append({
            "id": "yong-shen-state",
            "source_ref": f"chart.lines[{yong_line.get('position') - 1}]",
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
            "source_ref": "chart.yong_shen.fu_shen",
            "fact": yong_shen.get("fu_shen"),
            "conclusion_scope": CONCLUSION_SCOPE,
        })

    patterns = [
        _with_id(item, f"pattern-{index + 1}")
        for index, item in enumerate(chart.get("patterns", []))
    ]
    moving_relations = [
        _with_id(item, f"moving-relation-{item.get('position')}")
        for item in analysis.get("dong_yao_relations", [])
    ]
    hidden_lines = [
        _with_id(item, f"hidden-line-{index + 1}")
        for index, item in enumerate(analysis.get("hidden_lines", []))
    ]
    branch_relation_facts = [
        _with_id(
            item,
            f"branch-relation-{item.get('left_position')}-{item.get('right_position')}",
        )
        for item in analysis.get("branch_relation_facts", [])
    ]
    day_clash_facts = [
        _with_id(item, f"day-clash-{item.get('position')}")
        for item in analysis.get("day_clash_facts", [])
    ]
    moving_transformations = [
        _with_id(item, f"moving-transformation-{item.get('position')}")
        for item in analysis.get("moving_transformations", [])
    ]
    san_he_candidates = [
        _with_id(item, f"san-he-{item.get('wuxing')}-{index + 1}")
        for index, item in enumerate(analysis.get("san_he_candidates", []))
    ]
    yong_shen_force_chain = analysis.get("yong_shen_force_chain")
    if isinstance(yong_shen_force_chain, dict):
        yong_shen_force_chain = _with_id(yong_shen_force_chain, "force-chain")
    if yong_shen_force_chain:
        primary.append({
            "id": "force-chain",
            "source_ref": "analysis.yong_shen_force_chain",
            "fact": yong_shen_force_chain,
            "conclusion_scope": CONCLUSION_SCOPE,
        })

    for relation in moving_relations:
        primary.append({
            "id": f"moving-relation-{relation.get('position')}",
            "source_ref": f"analysis.moving_relations[{relation.get('position')}]",
            "fact": relation,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for pattern in patterns:
        secondary.append({
            "id": pattern["id"],
            "source_ref": f"chart.patterns[{pattern['id']}]",
            "fact": pattern,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for fact in day_clash_facts:
        secondary.append({
            "id": f"day-clash-{fact.get('position')}",
            "source_ref": f"analysis.day_clash_facts[{fact.get('position')}]",
            "fact": fact,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for fact in moving_transformations:
        primary.append({
            "id": f"moving-transformation-{fact.get('position')}",
            "source_ref": f"analysis.moving_transformations[{fact.get('position')}]",
            "fact": fact,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for fact in hidden_lines:
        secondary.append({
            "id": fact["id"],
            "source_ref": f"analysis.hidden_lines[{fact['id']}]",
            "fact": fact,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for fact in branch_relation_facts:
        branch_reference = {
            "id": fact["id"],
            "source_ref": f"analysis.branch_relation_facts[{fact['id']}]",
            "fact": fact,
            "conclusion_scope": CONCLUSION_SCOPE,
        }
        if "yong_shen" in fact.get("targets", []) or "world" in fact.get("targets", []):
            secondary.append(branch_reference)
        else:
            reference_only_facts.append({
                **fact,
                "reason": "未直接作用用神或世爻；仅作参考",
            })
    for fact in san_he_candidates:
        target = secondary if fact.get("targets") else None
        san_he_reference = {
            "id": fact["id"],
            "source_ref": f"analysis.san_he_candidates[{fact['id']}]",
            "fact": fact,
            "conclusion_scope": "structure_candidate_only_not_outcome",
        }
        if target is not None:
            target.append(san_he_reference)
        else:
            reference_only_facts.append({**fact, "reason": "无目标关联；仅作结构参考"})

    reference.append({
        "id": "hexagram-name",
        "source_ref": "chart.gua_name",
        "fact": {"gua_name": chart.get("gua_name")},
        "conclusion_scope": "reference_not_primary_judgment",
    })
    for line in lines:
        reference.append({
            "id": f"liu-shou-{line.get('position')}",
            "source_ref": f"chart.lines[{line.get('position') - 1}].liu_shou",
            "fact": {"position": line.get("position"), "liu_shou": line.get("liu_shou")},
            "conclusion_scope": "reference_not_primary_judgment",
        })
    gua_ci = chart.get("gua_ci")
    if gua_ci:
        gua_ci = {
            "name": gua_ci.get("name", ""),
            "text": gua_ci.get("text", ""),
            "yao_ci": gua_ci.get("yao_ci", []),
        }
        reference.append({
            "id": "gua-ci",
            "source_ref": "chart.gua_ci",
            "fact": gua_ci,
            "conclusion_scope": "reference_not_primary_judgment",
        })

    try:
        validated_conflicts = TypeAdapter(list[DivinationConflict]).validate_python(
            analysis.get("conflicts", [])
        )
    except ValidationError as exc:
        raise ValueError(f"liuyao analysis conflicts are invalid: {exc}") from exc
    conflicts = [item.model_dump() for item in validated_conflicts]

    evidence = {
        "primary": primary,
        "secondary": secondary,
        "reference": reference,
        "conflicts": conflicts,
    }

    analysis_output = {
        "moving_relations": moving_relations,
        "hidden_lines": hidden_lines,
        "branch_relation_facts": branch_relation_facts,
        "day_clash_facts": day_clash_facts,
        "moving_transformations": moving_transformations,
        "san_he_candidates": san_he_candidates,
        "timing_candidates": analysis.get("timing_candidates", []),
        "yong_shen_force_chain": yong_shen_force_chain,
        "yong_shen_candidates": [
            _with_id(item, f"yongshen-candidate-{item.get('position')}")
            for item in analysis.get("yong_shen_candidates", [])
        ],
        "conflicts": conflicts,
        "tomb_school": analysis.get("tomb_school"),
        "reference_only_facts": reference_only_facts,
    }

    return {
        "question": question,
        "casting": casting_projection(casting),
        "liuyao_chart": chart,
        "evidence": evidence,
        "analysis": analysis_output,
        "policy": {
            "may_output": ["tendency", "condition", "timing_window", "action"],
            "must_cover_conflicts": True,
            "forbidden": ["absolute_guarantee"],
        },
    }


def casting_projection(casting: dict) -> dict:
    projected = {
        "mode": casting.get("mode"),
        "order": casting.get("order", "bottom_up"),
        "yaos": casting.get("yaos", []),
        "dong_yao": casting.get("dong_yao", []),
    }
    if projected["mode"] == "coins":
        projected["coin_convention"] = casting.get("coin_convention")
        projected["rounds"] = casting.get("rounds", [])
    return projected
