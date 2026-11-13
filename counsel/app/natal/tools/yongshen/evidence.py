"""Projections for yongshen context, evidence, and the public decision."""

from .context import candidate

PUBLISHED_FIELDS = (
    "selected",
    "status",
    "conflicts",
    "rule_id",
    "ge_ju_context",
    "evidence",
    "policy_id",
)


def ge_ju_context(facts: dict, derived: dict, policy: dict) -> dict:
    ge_ju = facts.get("ge_ju") or {}
    themes = policy["tables"].get("ge_ju_themes") or {}
    raw_ten_god = ge_ju.get("pattern_god_ten_god") or ""
    ten_god = raw_ten_god or "杂格"
    return {
        "pattern": ge_ju.get("ge_ju", ""),
        "pattern_god_ten_god": raw_ten_god or "无格神",
        "pattern_god_source": ge_ju.get("pattern_god_source", ""),
        "yong_fa": ge_ju.get("yong_fa", ""),
        "is_clear": derived.get("ge_ju_clear", False),
        "theme": themes.get(ten_god, themes.get("杂格", "")),
    }


def evidence(facts: dict, derived: dict, policy: dict) -> list[dict]:
    return [
        {
            "source": "facts",
            "month_zhi": facts.get("month_zhi", ""),
            "day_master_stem": facts.get("day_master_stem", ""),
            "day_master_element": facts.get("day_master_element", ""),
            "season_is_extreme": derived.get("season_is_extreme", False),
        },
        _fu_yi_evidence(facts),
        _tiao_hou_evidence(facts, derived),
        _ge_ju_evidence(facts, derived),
        _cong_ge_evidence(facts, derived),
    ]


def _fu_yi_evidence(facts: dict) -> dict:
    fu_yi = facts.get("fu_yi") or {}
    basis = fu_yi.get("basis") or {}
    return {
        "source": "fu_yi",
        "qiang_ruo": fu_yi.get("qiang_ruo", ""),
        "pattern": fu_yi.get("pattern", ""),
        "ri_gan_gen": basis.get("ri_gan_gen", ""),
        "season": basis.get("season", ""),
        "yin_bi_shu_liang": basis.get("yin_bi_shu_liang", 0),
        "yong": candidate(fu_yi.get("yong")),
        "xi": candidate(fu_yi.get("xi")),
        "ji": candidate(fu_yi.get("ji")),
    }


def _tiao_hou_evidence(facts: dict, derived: dict) -> dict:
    tiao_hou = facts.get("tiao_hou") or {}
    return {
        "source": "tiao_hou",
        "primary_wuxing": tiao_hou.get("primary_wuxing", ""),
        "secondary_wuxing": candidate(tiao_hou.get("secondary_wuxing")),
        "secondary_condition": tiao_hou.get("secondary_condition", ""),
        "condition_ok": derived.get("tiaohou_secondary_condition_ok", False),
        "secondary_usable": derived.get("tiaohou_secondary_usable", False),
        "primary_present": derived.get("tiaohou_primary_present", False),
        "primary_visible": derived.get("tiaohou_primary_visible", False),
        "secondary_present": derived.get("tiaohou_secondary_present", False),
        "secondary_visible": derived.get("tiaohou_secondary_visible", False),
        "yong": candidate(derived.get("tiaohou_yong")),
        "yong_source": derived.get("tiaohou_yong_source", "none"),
        "yong_present": derived.get("tiaohou_yong_present", False),
        "yong_visible": derived.get("tiaohou_yong_visible", False),
        "control_wuxing": candidate(derived.get("tiaohou_control_wuxing")),
    }


def _ge_ju_evidence(facts: dict, derived: dict) -> dict:
    ge_ju = facts.get("ge_ju") or {}
    return {
        "source": "ge_ju",
        "pattern": ge_ju.get("ge_ju", ""),
        "pattern_god_ten_god": ge_ju.get("pattern_god_ten_god") or "无格神",
        "pattern_god_source": ge_ju.get("pattern_god_source", ""),
        "is_clear": derived.get("ge_ju_clear", False),
    }


def _cong_ge_evidence(facts: dict, derived: dict) -> dict:
    fu_yi = facts.get("fu_yi") or {}
    return {
        "source": "cong_ge",
        "pattern": fu_yi.get("pattern", ""),
        "is_verified": derived.get("congge_verified", False),
        "is_fake_candidate": derived.get("congge_fake", False),
        "guan_sha_effective_count": derived.get("guan_sha_effective_count", 0),
        "cai_effective_count": derived.get("cai_effective_count", 0),
        "shi_shang_effective_count": derived.get("shi_shang_effective_count", 0),
        "yin_rooted_count": derived.get("yin_rooted_count", 0),
        "bi_jie_rooted_count": derived.get("bi_jie_rooted_count", 0),
    }


def published_yongshen(result: dict) -> dict:
    """Project the internal resolver result to the MCP decision object."""
    published = {field: result[field] for field in PUBLISHED_FIELDS}
    published["selected_yongshen"] = published.pop("selected")
    return published
