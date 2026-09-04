"""测试共用 helper：tools 路径注入 + mock 因子快照。"""
import os
import sys

# 注入 tools 目录到 sys.path（tests 与 tools 平级于 skills/liki-bazi 下）
TOOLS = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'skills', 'liki-bazi', 'tools')
if TOOLS not in sys.path:
    sys.path.insert(0, TOOLS)


def mock_base_context(**shishen):
    """构造最小因子快照（base）。"""
    base = {"shishen": shishen, "ri_gan": "甲",
           "wuxing": {"wang_shuai": {}}}
    return base


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
