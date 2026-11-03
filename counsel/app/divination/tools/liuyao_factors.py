"""六爻 case 因子投影；搬运 engine 事实并做解释层级组织，不新增命理判断。"""

from pydantic import TypeAdapter, ValidationError

from app.contracts.divination import DivinationConflict

CONCLUSION_SCOPE = "conditional_candidate_not_outcome"


def _with_id(item: dict, fact_id: str) -> dict:
    return {"id": fact_id, **item}


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

    def project_line(line: dict) -> dict:
        return {
            "position": line.get("position"),
            "yao_value": line.get("yao_value"),
            "gan_zhi": f"{line.get('gan', '')}{line.get('zhi', '')}",
            "wuxing": line.get("wuxing"),
            "liu_qin": line.get("liu_qin"),
            "liu_shou": line.get("liu_shou"),
            "shi_ying": line.get("shi_ying"),
            "chang_sheng_yue": line.get("chang_sheng_yue"),
            "wang_shuai": line.get("wang_shuai"),
            "ri_chen_relations": line.get("ri_chen_relations"),
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
        }

    lines = []
    wang_shuai = chart.get("wang_shuai", []) or []
    day_relations = chart.get("ri_chen_relations", []) or []
    for index, line in enumerate(chart.get("lines", [])):
        if not isinstance(line, dict):
            raise ValueError("chart lines contains a non-object")
        lines.append(project_line({
            **line,
            "wang_shuai": wang_shuai[index] if index < len(wang_shuai) else None,
            "ri_chen_relations": (
                day_relations[index] if index < len(day_relations) else None
            ),
        }))

    bian_lines = []
    for line in chart.get("bian_yao", []) or []:
        if not isinstance(line, dict):
            raise ValueError("chart bian_yao contains a non-object")
        bian_lines.append(project_line(line))

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
    reference_only_facts = []

    if yong_line:
        primary.append({
            "id": "yong-shen-state",
            "source_ref": f"liuyao_chart.lines[{yong_line.get('position') - 1}]",
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
            "source_ref": "selected_yongshen_context.selected_yongshen.fu_shen",
            "fact": yong_shen.get("fu_shen"),
            "conclusion_scope": CONCLUSION_SCOPE,
        })

    patterns = [
        {
            **_with_id(item, f"pattern-{index + 1}"),
            "kind": item.get("pattern_kind", ""),
        }
        for index, item in enumerate(chart.get("ge_ju", []))
    ]
    moving_relations = [
        _with_id(item, f"moving-relation-{item.get('position')}")
        for item in chart.get("dong_yao_relations", [])
    ]
    hidden_lines = [
        _with_id(item, f"hidden-line-{index + 1}")
        for index, item in enumerate(chart.get("hidden_lines", []))
    ]
    branch_relation_facts = [
        _with_id(
            item,
            f"branch-relation-{item.get('left_position')}-{item.get('right_position')}",
        )
        for item in chart.get("branch_relation_facts", [])
    ]
    day_clash_facts = [
        _with_id(item, f"day-clash-{item.get('position')}")
        for item in chart.get("day_clash_facts", [])
    ]
    moving_transformations = [
        _with_id(item, f"moving-transformation-{item.get('position')}")
        for item in chart.get("moving_transformations", [])
    ]
    san_he_candidates = [
        _with_id(
            item,
            f"san-he-{item.get('wuxing')}-{index + 1}",
        )
        for index, item in enumerate(chart.get("san_he_candidates", []))
    ]
    yong_shen_force_chain = chart.get("yong_shen_force_chain")
    if isinstance(yong_shen_force_chain, dict):
        yong_shen_force_chain = _with_id(yong_shen_force_chain, "force-chain")
    if yong_shen_force_chain:
        primary.append({
            "id": "force-chain",
            "source_ref": "chart_facts.yong_shen_force_chain",
            "fact": yong_shen_force_chain,
            "conclusion_scope": CONCLUSION_SCOPE,
        })

    for relation in moving_relations:
        primary.append({
            "id": f"moving-relation-{relation.get('position')}",
            "source_ref": f"chart_facts.moving_relations[{relation.get('position')}]",
            "fact": relation,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for pattern in patterns:
        secondary.append({
            "id": pattern["id"],
            "source_ref": f"liuyao_chart.patterns[{pattern['id']}]",
            "fact": pattern,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for fact in day_clash_facts:
        secondary.append({
            "id": f"day-clash-{fact.get('position')}",
            "source_ref": f"chart_facts.day_clash_facts[{fact.get('position')}]",
            "fact": fact,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for fact in moving_transformations:
        primary.append({
            "id": f"moving-transformation-{fact.get('position')}",
            "source_ref": f"chart_facts.moving_transformations[{fact.get('position')}]",
            "fact": fact,
            "conclusion_scope": CONCLUSION_SCOPE,
        })

    for fact in hidden_lines:
        secondary.append({
            "id": fact["id"],
            "source_ref": f"chart_facts.hidden_lines[{fact['id']}]",
            "fact": fact,
            "conclusion_scope": CONCLUSION_SCOPE,
        })
    for fact in branch_relation_facts:
        branch_reference = {
            "id": fact["id"],
            "source_ref": f"chart_facts.branch_relation_facts[{fact['id']}]",
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
            "source_ref": f"chart_facts.san_he_candidates[{fact['id']}]",
            "fact": fact,
            "conclusion_scope": "structure_candidate_only_not_outcome",
        }
        if target is not None:
            target.append(san_he_reference)
        else:
            reference_only_facts.append({**fact, "reason": "无目标关联；仅作结构参考"})

    reference.append({
        "id": "hexagram-name",
        "source_ref": "liuyao_chart.gua_name",
        "fact": {"gua_name": chart.get("gua_name")},
        "conclusion_scope": "reference_not_primary_judgment",
    })
    for line in lines:
        reference.append({
            "id": f"liu-shou-{line.get('position')}",
            "source_ref": f"liuyao_chart.lines[{line.get('position') - 1}].liu_shou",
            "fact": {"position": line.get("position"), "liu_shou": line.get("liu_shou")},
            "conclusion_scope": "reference_not_primary_judgment",
        })
    gua_ci = chart.get("gua_ci")
    if gua_ci:
        gua_ci = {
            "name": gua_ci.get("name", ""),
            "text": gua_ci.get("gua_ci", ""),
            "yao_ci": gua_ci.get("yao_ci", []),
        }
        reference.append({
            "id": "gua-ci",
            "source_ref": "liuyao_chart.gua_ci",
            "fact": gua_ci,
            "conclusion_scope": "reference_not_primary_judgment",
        })

    try:
        validated_conflicts = TypeAdapter(list[DivinationConflict]).validate_python(
            chart.get("conflicts", [])
        )
    except ValidationError as exc:
        raise ValueError(f"liuyao chart conflicts are invalid: {exc}") from exc
    conflicts = [item.model_dump() for item in validated_conflicts]

    evidence = {
        "primary": primary,
        "secondary": secondary,
        "reference": reference,
        "conflicts": conflicts,
    }

    timing_candidates = chart.get("timing_candidates")
    if not isinstance(timing_candidates, list):
        timing_candidates = []

    return {
        "question": question,
        "casting": casting_projection(casting),
        "liuyao_chart": {
            "gua_name": chart.get("gua_name"),
            "bian_gua": chart.get("bian_gua"),
            "palace": chart.get("gong"),
            "palace_wuxing": chart.get("gong_wuxing"),
            "lines": lines,
            "bian_lines": bian_lines,
            "ri_gan": chart.get("ri_chen_gan"),
            "ri_zhi": chart.get("ri_chen_zhi"),
            "yue_gan": chart.get("yue_jian_gan"),
            "yue_zhi": chart.get("yue_jian_zhi"),
            "xun_kong": chart.get("xun_kong", []),
            "patterns": patterns,
            "gua_ci": {
                "name": gua_ci.get("name", ""),
                "text": gua_ci.get("gua_ci", ""),
                "yao_ci": gua_ci.get("yao_ci", []),
            } if gua_ci else None,
        },
        "selected_yongshen_context": {
            "selected_yongshen": yong_shen,
            "yong_line": yong_line,
            "world_line": world,
            "other_line": other,
        },
        "evidence": evidence,
        "chart_facts": {
            "moving_relations": moving_relations,
            "patterns": patterns,
            "hidden_lines": hidden_lines,
            "branch_relation_facts": branch_relation_facts,
            "day_clash_facts": day_clash_facts,
            "moving_transformations": moving_transformations,
            "san_he_candidates": san_he_candidates,
            "yong_shen_force_chain": yong_shen_force_chain,
            "yong_shen_candidates": [
                _with_id(item, f"yongshen-candidate-{item.get('position')}")
                for item in chart.get("yong_shen_candidates", [])
            ],
            "reference_only_facts": reference_only_facts,
        },
        "timing_candidates": timing_candidates,
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
        "casting_digest": casting.get("casting_digest"),
    }
    if projected["mode"] == "coins":
        projected["rounds"] = casting.get("rounds", [])
    return projected
