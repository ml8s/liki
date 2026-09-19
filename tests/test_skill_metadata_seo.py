"""Skill 元数据 SEO 关键词锁定：确保描述覆盖核心搜索短语，防止回潮。"""
from pathlib import Path

import yaml

SKILL_MD = Path(__file__).resolve().parents[1] / "skills/liki/SKILL.md"

REQUIRED_PHRASES = [
    # 八字核心
    "八字算命", "生辰八字算命", "排八字", "看八字", "四柱命盘",
    "紫微斗数", "紫微命盘", "大运流年", "命盘分析",
    # 场景
    "婚姻分析", "八字合婚", "婚姻配对", "情侣合盘", "感情复合",
    "事业分析", "职业方向", "官运", "考编",
    "财运分析", "偏财正财", "投资时机", "副业",
    "学业分析", "考试运", "性格分析", "外貌长相",
    "健康分析", "五行体质", "怀孕生育", "人际贵人",
    "官司纠纷", "房产运", "出行安全",
    # 六爻 / 问事
    "六爻占卜", "算卦", "摇卦起卦", "问事业", "问财运",
    "问感情", "问学业", "问失物", "问官司", "应期分析",
    # 奇门
    "奇门遁甲", "奇门问事", "策略分析", "谈判时机",
    # 择日
    "黄历择日", "选日子", "挑吉日", "老黄历",
    "结婚吉日", "开业吉日", "搬家吉日", "入宅择日",
    "装修择日", "动土吉日", "安床吉日", "安葬择日",
    # 风水
    "看风水", "家居风水", "风水布局", "家居布局",
    "办公室风水", "新房风水", "店铺选址", "商铺风水",
    "八宅风水", "命卦", "玄空风水", "玄空飞星", "流年飞星",
    # 起名
    "宝宝起名", "宝宝取名", "新生儿起名", "新生儿取名",
    "婴儿起名", "小孩取名", "成人改名", "公司起名",
    "品牌命名", "宠物取名", "外国人中文名", "名字测试", "名字评估",
    "八字起名", "五行起名",
    # 英文
    "BaZi chart", "Chinese astrology", "Four Pillars of Destiny",
    "Zi Wei Dou Shu", "I Ching divination", "Liu Yao",
    "Qi Men Dun Jia", "Chinese almanac", "date selection",
    "Feng Shui analysis", "home layout", "Chinese baby naming",
]


def _description() -> str:
    txt = SKILL_MD.read_text(encoding="utf-8")
    assert txt.startswith("---\n")
    meta = yaml.safe_load(txt.split("---\n")[1])
    assert isinstance(meta, dict)
    desc = meta.get("description") or ""
    assert desc.strip(), "SKILL.md description 不能为空"
    return desc


def test_display_name_contains_brand_and_category():
    txt = SKILL_MD.read_text(encoding="utf-8")
    meta = yaml.safe_load(txt.split("---\n")[1])
    display = meta.get("displayName") or ""
    assert "Liki" in display
    assert "命理" in display


def test_summary_contains_core_categories():
    txt = SKILL_MD.read_text(encoding="utf-8")
    meta = yaml.safe_load(txt.split("---\n")[1])
    summary = meta.get("summary") or ""
    for kw in ("八字", "紫微", "六爻", "奇门", "择日", "风水", "起名"):
        assert kw in summary, f"summary 缺少核心品类词: {kw}"


def test_description_covers_all_required_phrases():
    desc = _description()
    missing = [p for p in REQUIRED_PHRASES if p not in desc]
    assert not missing, f"description 缺少搜索短语: {missing}"


def test_description_has_compliance_boundary():
    desc = _description()
    assert "传统文化视角" in desc
    assert "不构成专业建议" in desc


def test_description_does_not_exaggerate():
    desc = _description()
    for bad in ("最准", "大师", "改运", "必中", "百分之百"):
        assert bad not in desc, f"description 含违规词: {bad}"
