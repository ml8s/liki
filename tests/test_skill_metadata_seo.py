"""Skill 元数据 SEO 契约：高频词独立命中、长尾精选、长度受控。"""
from pathlib import Path
import re
import yaml

SKILL_MD = Path(__file__).resolve().parents[1] / "skills/liki/SKILL.md"

P0_PHRASES = [
    "算命", "命理", "运势", "八字", "生辰八字", "排八字", "四柱",
    "紫微斗数", "紫微命盘", "大运", "流年", "六爻", "占卜", "算卦",
    "奇门遁甲", "黄历", "老黄历", "择日", "选日子", "风水", "家居风水",
    "风水布局", "起名", "取名", "宝宝起名", "宝宝取名",
]

P1_PHRASES = [
    "八字合婚", "婚姻分析", "感情走向", "事业分析", "职业方向",
    "财运分析", "投资时机", "学业分析", "考试运", "健康分析",
    "五行体质", "怀孕生育时机", "新生儿起名", "成人改名", "公司起名",
    "品牌命名", "名字测试", "八字起名", "结婚吉日", "开业吉日",
    "搬家吉日", "办公室风水", "店铺选址", "八宅风水", "玄空飞星",
]

ENGLISH_PHRASES = [
    "BaZi", "Chinese astrology", "Four Pillars of Destiny",
    "Zi Wei Dou Shu", "I Ching divination", "Chinese almanac",
    "Feng Shui", "Chinese baby naming",
]


def _metadata() -> dict:
    txt = SKILL_MD.read_text(encoding="utf-8")
    assert txt.startswith("---\n")
    meta = yaml.safe_load(txt.split("---\n")[1])
    assert isinstance(meta, dict)
    return meta


def _description() -> str:
    meta = _metadata()
    desc = meta.get("description") or ""
    assert desc.strip(), "SKILL.md description 不能为空"
    return desc


def test_display_name_contains_brand_and_category():
    meta = _metadata()
    display = meta.get("displayName") or ""
    assert "Liki" in display
    assert "命理" in display


def test_summary_contains_core_categories_and_naming_pair():
    summary = _metadata().get("summary") or ""
    for kw in ("八字", "紫微", "六爻", "奇门", "择日", "风水", "起名", "取名"):
        assert kw in summary, f"summary 缺少核心品类词: {kw}"
    assert "起名、取名" in summary


def test_description_has_controlled_length():
    desc = _description()
    assert len(desc) <= 1000, f"description 超长: {len(desc)}"
    assert len(desc) >= 450, f"description 过短，丢失核心能力: {len(desc)}"


def test_description_covers_all_p0_phrases():
    desc = _description()
    missing = [p for p in P0_PHRASES if p not in desc]
    assert not missing, f"description 缺少 P0 搜索词: {missing}"


def test_description_covers_selected_p1_phrases():
    desc = _description()
    missing = [p for p in P1_PHRASES if p not in desc]
    assert not missing, f"description 缺少精选搜索词: {missing}"


def test_naming_keywords_are_delimited_pair():
    desc = _description()
    # 起名 / 取名必须是相邻、分隔的能力词，不能只作为合成词的后缀。
    assert "支持起名、取名" in desc, "description 需要“起名、取名”独立关键词对"


def test_description_has_compact_english_terms():
    desc = _description()
    match = re.search(r"Also supports ([^。]+)", desc)
    assert match, "description 缺少英文搜索词"
    english = match.group(1)
    assert len(english) <= 150, f"英文关键词过长: {len(english)}"
    missing = [p for p in ENGLISH_PHRASES if p not in english]
    assert not missing, f"英文关键词缺失: {missing}"


def test_description_has_compliance_boundary():
    desc = _description()
    assert "传统文化视角" in desc
    assert "不构成专业建议" in desc


def test_description_does_not_exaggerate():
    desc = _description()
    for bad in ("最准", "大师", "改运", "必中", "百分之百"):
        assert bad not in desc, f"description 含违规词: {bad}"
