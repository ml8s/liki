"""测试共用 helper：tools 路径注入 + mock 因子快照。"""
import os
import sys

# 注入 tools 目录到 sys.path（tests 与 tools 平级于 skills/liki-bazi 下）
TOOLS = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'skills', 'liki-bazi', 'tools')
if TOOLS not in sys.path:
    sys.path.insert(0, TOOLS)


def mock_factors(**shishen):
    """构造最小因子快照（fac）。"""
    fac = {"shishen": shishen, "ri_gan": "甲", "qiangruo": "中和",
           "wuxing": {"wang_shuai": {}}}
    return fac
