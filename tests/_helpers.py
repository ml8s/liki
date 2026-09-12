"""测试共用 helper：tools 路径注入 + mock 因子快照。"""
import os
import sys

# 注入 tools 目录到 sys.path（tests 与 tools 平级于 skills/liki-bazi 下）
TOOLS = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'skills', 'liki-bazi', 'tools')
if TOOLS not in sys.path:
    sys.path.insert(0, TOOLS)


def mock_base_context(**ten_god_states):
    """构造最小因子快照（base）。"""
    base = {
        "ten_god_states": ten_god_states,
        "element_states": {},
        "ri_gan": "甲",
        "wuxing": {"count": {}},
    }
    return base


def mock_engine_facts() -> dict:
    """构造 fullchart engine 事实的最小契约片段。"""
    return {
        "lu_roots": [],
        "relation_groups": [],
        "ten_god_states": [{
            "shi_shen": "比肩", "wuxing": "木", "transparent": True, "hidden": True,
            "rooted": True, "timely": True, "count": 1, "strength": "strong",
        }],
        "element_states": [
            {
                "wuxing": element, "season_strength": season, "strength": season,
                "transparent": element == "木", "rooted": element == "木",
                "controls": controls, "controlled_by": controlled_by,
                "controller_strength": "weak", "reasons": ["timely"] if season == "strong" else [],
            }
            for element, season, controls, controlled_by in (
                ("木", "strong", "土", "金"),
                ("火", "strong", "金", "水"),
                ("土", "weak", "水", "木"),
                ("金", "weak", "木", "火"),
                ("水", "weak", "火", "土"),
            )
        ],
        "atomic_facts": {
            "day_master_element": "木", "officer_killing_cleaned": False,
            "day_master_stem": "甲", "day_branch": "子",
            "wealth_tomb_present": False, "wealth_star_in_tomb": False,
            "spouse_palace_state": "静", "day_branch_type": "桃花",
            "year_officer_killing": False,
            "month_longevity": "沐浴", "year_stem_ten_god": "比肩",
            "month_main_ten_god": "正印", "hour_stem_ten_god": "比肩",
            "pattern_god_transparent": False,
            "pillar_punishments": {"nian": False, "yue": False, "ri": False, "shi": False},
        },
        "da_yun": {"steps": [{"rooted": False, "root_refs": []}]},
    }


def mock_yong_shen() -> dict:
    """构造 bazi.fullchart 用神三派的最小契约片段。"""
    return {
        "fu_yi": {
            "wuxing_count": {"木": 1, "火": 1, "土": 1, "金": 1, "水": 1},
            "wang_shuai": {"木": "旺", "火": "相", "土": "死", "金": "囚", "水": "休"},
            "yong": "木", "xi": "水", "ji": "金", "qiangruo": "身强",
        },
        "tiao_hou": {"yong": "火", "xi": "木", "ji": "水", "season": "春", "detail": "mock"},
        "ge_ju": {"yong": "木", "xi": "水", "ji": "金", "ge_ju": "正印格", "yong_fa": "顺用"},
    }


def mock_ziwei() -> dict:
    """构造 factor 层需要的最小紫微原子事实盘。"""
    palaces = (
        "命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄",
        "迁移", "仆役", "官禄", "田宅", "福德", "父母",
    )
    return {
        "gong_wei": [{"name": palace} for palace in palaces],
        "palace_facts": [
            {"palace": "命宫", "kind": "special", "target": "无主星"}
        ],
    }


def valid_daxian(birth_year: int = 1990):
    """构造引擎 ziwei.daxian 形状的最小 12 段大限。"""
    palaces = (
        "命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄",
        "迁移", "仆役", "官禄", "田宅", "福德", "父母",
    )
    return [
        {
            "gong": palace,
            "name": f"限{i + 1}",
            "start_year": birth_year + i * 10,
            "end_year": birth_year + i * 10 + 9,
            "qi_sui": i * 10 + 1,
            "zhi_sui": i * 10 + 10,
        }
        for i, palace in enumerate(palaces)
    ]
