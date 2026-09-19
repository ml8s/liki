"""命理前置条件契约：候选断语必须携带喜忌、性别与婚姻状态门槛。"""
import csv
from pathlib import Path

TOOLS = Path(__file__).resolve().parents[1] / "skills/liki/bazi/tools"

def _load(path: str) -> list[dict]:
    with (TOOLS / path).open(encoding="utf-8-sig", newline="") as fh:
        return list(csv.DictReader(fh))

def _condition_groups(assertion_id: str) -> dict[int, set[tuple[str, str]]]:
    groups: dict[int, set[tuple[str, str]]] = {}
    for row in _load("assertions/assertion_conditions.csv"):
        if row["assertion_id"] == assertion_id:
            groups.setdefault(int(row["condition_group_id"]), set()).add((row["factor"], row["expected"]))
    return groups

def _condition_names(assertion_id: str) -> dict[int, set[str]]:
    return {group: {factor for factor, _ in conditions}
            for group, conditions in _condition_groups(assertion_id).items()}

def _assertion(assertion_id: str) -> dict:
    return next(row for row in _load("assertions/assertions.csv")
                if row["assertion_id"] == assertion_id)

def test_female_marriage_breakdown_requires_female_and_injured_officer() -> None:
    """婚变候选不得用于男命，也不得用粗粒度食伤旺替代伤官克官。"""
    factors = _condition_groups("ying_h19")
    names = _condition_names("ying_h19")
    assert factors and set(factors) == set(range(1, 8))
    for group in factors.values():
        assert ("性别", "female") in group
        assert ("本命伤官克官", "1") in group
        assert ("本命婚姻不稳", "1") in group
        assert ("本命食伤旺", "1") not in group
    assert all("流年配偶星透" in names[group] for group in names)

def test_marriage_start_requires_favorable_star_and_excludes_annual_damage() -> None:
    """配偶星显动只有为用且不受克时，才可作成婚候选。"""
    factors = _condition_groups("ymar_103")
    assert set(factors) == {1}
    assert factors[1] >= {
        ("流年配偶星透", "1"), ("流年配偶星大运窗口", "1"),
        ("本命配偶星为用", "1"), ("流年配偶星克", "0"),
    }
    conditions = _load("assertions/assertion_conditions.csv")
    expected = {
        ("流年配偶星透", "1"), ("流年配偶星大运窗口", "1"),
        ("本命配偶星为用", "1"), ("流年配偶星克", "0"),
    }
    actual = {(r["factor"], r["expected"]) for r in conditions
              if r["assertion_id"] == "ymar_103"}
    assert expected <= actual

def test_spouse_star_damage_only_counts_when_star_is_favorable() -> None:
    """配偶星为忌受制不是婚变；为用受克才是受损候选。"""
    factors = _condition_groups("ymar_108")
    assert set(factors) == {1}
    assert {("流年配偶星克", "1"), ("本命配偶星为用", "1")} <= factors[1]

def test_wealth_signals_separate_opportunity_from_result() -> None:
    """财透只是显动；得财必须身强任财且财为用。"""
    assert "得财与否须身强任财且财为用" in _assertion("ycai_103")["结论"]
    ycai_121 = _condition_groups("ycai_121")
    assert set(ycai_121) == {1}
    assert {("流年财星透", "1"), ("本命身强", "1"), ("本命财星为用", "1")} <= ycai_121[1]
    ycai_122 = _condition_groups("ycai_122")
    assert set(ycai_122) == {1}
    assert {("流年财星克", "1"), ("本命财星为用", "1")} <= ycai_122[1]

def test_generic_changes_do_not_prescribe_answer_direction() -> None:
    """天克地冲不能按题目反向定得财/破财或成婚/婚变。"""
    assert "按题指认" not in _assertion("ycai_108")["结论"]
    assert "按诉求" not in _assertion("ymar_112")["结论"]
    assert "由财星喜忌" in _assertion("ycai_108")["结论"]
    assert "方向须由喜忌" in _assertion("ymar_112")["结论"]


def _factor_groups(factor_id: str) -> dict[int, list[dict[str, str]]]:
    groups: dict[int, list[dict[str, str]]] = {}
    for row in _load("factors/factors.csv"):
        if row["factor_id"] == factor_id:
            groups.setdefault(int(row["group_id"]), []).append(row)
    return groups

def test_female_marriage_instability_is_gender_gated() -> None:
    """女命婚姻不稳不能因命主为男而命中。"""
    groups = _factor_groups("女命婚姻不稳")
    assert set(groups) == {1, 2, 3}
    for group in groups.values():
        terms = {(row["kind"], row["expression"], row["expected"]) for row in group}
        assert ("condition", "直读[gender,female]", "1") in terms

def test_marriage_stability_branches_are_gender_gated() -> None:
    """男命忌比劫夺财、女命忌伤官克官；两条稳定性分支不得混用性别。"""
    groups = _factor_groups("婚姻基础稳定")
    assert set(groups) == {1, 2}
    male_terms = {(r["expression"], r["expected"]) for r in groups[1]}
    female_terms = {(r["expression"], r["expected"]) for r in groups[2]}
    assert ("直读[gender,male]", "1") in male_terms
    assert ("比劫夺财", "0") in male_terms
    assert ("直读[gender,female]", "1") in female_terms
    assert ("伤官克官", "0") in female_terms

def test_marriage_shensha_are_candidates_not_verdicts() -> None:
    """红鸾、天喜、桃花与日支神煞只能作取象候选。"""
    for aid in ("ymar_115", "ymar_116", "ymar_117", "hun_404", "hun_405", "hun_406"):
        row = _assertion(aid)
        assert "候选" in row["结论"]
        assert "不单独定" in row["结论"] or "须" in row["结论"]

def test_unreachable_spouse_palace_contradiction_is_removed() -> None:
    """engine 夫妻宫状态是优先级归并后的单值，冲与合不可能同时命中。"""
    assert not _condition_groups("hun_412")
    assert all(row["assertion_id"] != "hun_412" for row in _load("assertions/assertions.csv"))

def _condition_rows(aid):
    return [r for r in _load("assertions/assertion_conditions.csv") if r["assertion_id"] == aid]

def test_ziwei_positive_spouse_star_requires_support_and_lack_of_evil():
    """吉性夫妻宫取象只在主星得支持、无煞无忌时命中。"""
    for aid, star in {
        "zf_305": "夫妻宫天同", "zf_307": "夫妻宫天府",
        "zf_308": "夫妻宫太阴", "zf_311": "夫妻宫天相",
    }.items():
        groups = _condition_groups(aid)
        assert set(groups) == {1, 2}
        for group in groups.values():
            assert (star, "1") in group
            assert ("夫妻宫无煞", "1") in group
            assert ("夫妻宫化忌", "0") in group
            bright = star + "庙旺"
            supported = (bright, "1") in group or ("夫妻宫六吉星", "1") in group
            assert supported

def test_ziwei_adverse_spouse_star_requires_adverse_modifier():
    """波折/变动取象不得仅凭主星命中，必须叠落陷、煞星或化忌。"""
    cases = {
        "zf_302": "夫妻宫天机", "zf_304": "夫妻宫武曲",
        "zf_306": "夫妻宫廉贞", "zf_309": "夫妻宫贪狼",
        "zf_310": "夫妻宫巨门", "zf_313": "夫妻宫七杀",
        "zf_314": "夫妻宫破军",
    }
    modifiers = {"夫妻宫煞星", "夫妻宫化忌"}
    for aid, star in cases.items():
        groups = _condition_groups(aid)
        assert groups
        for group in groups.values():
            assert (star, "1") in group
            hit_modifiers = modifiers & {factor for factor, _ in group}
            if hit_modifiers:
                assert len(group) == 2
            else:
                fallen = star + "落陷"
                assert (fallen, "1") in group
                assert len(group) == 2

def test_spouse_palace_modifier_factors_are_engine_projected():
    """煞吉约束必须读取 engine palace facts，不在 Python 复算。"""
    expected = {
        "夫妻宫煞星": "宫含[夫妻,煞星,任意]",
        "夫妻宫六吉星": "宫含[夫妻,紫微六吉星,任意]",
    }
    rows = _load("factors/factors.csv")
    by_id = {r["factor_id"]: r for r in rows}
    for factor, expression in expected.items():
        assert by_id[factor]["expression"] == expression
        assert by_id[factor]["kind"] == "condition"

    star_brightness = {
        "夫妻宫天同庙旺": ("天同", "庙旺"),
        "夫妻宫天府庙旺": ("天府", "庙旺"),
        "夫妻宫太阴庙旺": ("太阴", "庙旺"),
        "夫妻宫天相庙旺": ("天相", "庙旺"),
        "夫妻宫天机落陷": ("天机", "落陷"),
        "夫妻宫武曲落陷": ("武曲", "落陷"),
        "夫妻宫廉贞落陷": ("廉贞", "落陷"),
        "夫妻宫贪狼落陷": ("贪狼", "落陷"),
        "夫妻宫巨门落陷": ("巨门", "落陷"),
        "夫妻宫七杀落陷": ("七杀", "落陷"),
        "夫妻宫破军落陷": ("破军", "落陷"),
    }
    for factor, (star, brightness) in star_brightness.items():
        assert by_id[factor]["expression"] == f"宫含[夫妻,{star},{brightness}]"
        assert by_id[factor]["kind"] == "condition"

def test_rooted_spouse_star_can_still_be_controlled():
    """有根配偶星可同时受制；不得把“得地”与“受克”定义为互斥。"""
    groups = _factor_groups("配偶星受克")
    assert set(groups) == {1}
    terms = {(r["kind"], r["expression"], r["expected"]) for r in groups[1]}
    assert ("condition", "克者旺[配偶星]", "1") in terms
    assert ("condition", "现[配偶星]", "1") in terms
    assert not any(kind == "factor_ref" and expression == "配偶星得地" and expected == "0"
                   for kind, expression, expected in terms)

    conditions = _condition_groups("hun_207")
    assert set(conditions) == {1}
    assert {("配偶星受克", "1"), ("配偶星得地", "1"), ("夫妻宫破", "1")} <= conditions[1]

def test_wealth_states_split_competition_from_seizure():
    """比劫争财是结构候选；弱财被劫才可断破财无蓄；得令之财被争为有财难守。"""
    competition = _factor_groups("比劫争财")
    assert set(competition) == {1}
    assert {("factor_ref", "比劫旺", "1"), ("condition", "现[财星]", "1")} <= {
        (r["kind"], r["expression"], r["expected"]) for r in competition[1]
    }

    seizure = _factor_groups("比劫夺财")
    assert set(seizure) == {1}
    assert {("factor_ref", "比劫旺", "1"), ("factor_ref", "财星弱", "1"), ("factor_ref", "财星现", "1")} <= {
        (r["kind"], r["expression"], r["expected"]) for r in seizure[1]
    }

    rooted = _condition_groups("cai_106")
    assert set(rooted) == {1}
    assert ("比劫争财", "1") in rooted[1]
    assert ("财星得令", "1") in rooted[1]
    assert ("财星弱", "0") in rooted[1]

    rootless = _condition_groups("cai_108")
    assert set(rootless) == {1}
    assert ("财星透根", "0") in rootless[1]
    assert ("弱财被劫", "1") in rootless[1]

def test_hidden_rooted_wealth_requires_hidden_root_and_no_peer_competition():
    """“财得地藏支”必须同时有藏支和有根，不得用得令替代得地。"""
    groups = _factor_groups("财星得地")
    assert set(groups) == {1}
    assert ("condition", "有根[财星]", "1") in {
        (r["kind"], r["expression"], r["expected"]) for r in groups[1]
    }

    conditions = _condition_groups("cai_103")
    assert set(conditions) == {1}
    assert ("财星藏", "1") in conditions[1]
    assert ("财星得地", "1") in conditions[1]
    assert ("财星透", "0") in conditions[1]
    assert ("比劫争财", "0") in conditions[1]
    assert ("财星得令", "1") not in conditions[1]
