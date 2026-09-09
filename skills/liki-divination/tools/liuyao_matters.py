"""六爻事项路由表；只做语义类别 → 默认用神，不做排盘或断卦。"""
from __future__ import annotations


MATTERS: dict[str, dict] = {
    "general": {
        "name": "一般问事",
        "yong_shen": "世爻",
        "notes": "问题不具体时只作当前状态参考；应先澄清具体目标。",
    },
    "self": {
        "name": "自身状态",
        "yong_shen": "世爻",
        "notes": "问自己当前处境、状态或承受能力。",
    },
    "other": {
        "name": "对方或环境",
        "yong_shen": "应爻",
        "notes": "问对方、外部环境或合作对象的状态。",
    },
    "career": {
        "name": "事业官运",
        "yong_shen": "官鬼",
        "notes": "工作、职务、面试、升迁、事业推进。",
    },
    "wealth": {
        "name": "求财生意",
        "yong_shen": "妻财",
        "notes": "收入、交易、投资结果、经营现金流。",
    },
    "relationship": {
        "name": "感情婚姻",
        "yong_shen": "",
        "notes": "必须先确认提问视角；不可把性别视为唯一口径。",
    },
    "study": {
        "name": "学业文书",
        "yong_shen": "父母",
        "notes": "考试、录取、证书、合同、文书。",
    },
    "home": {
        "name": "房产住宅",
        "yong_shen": "父母",
        "notes": "住宅、搬迁、房产买卖或产权手续。",
    },
    "travel": {
        "name": "出行",
        "yong_shen": "父母",
        "notes": "行程成行、途中状态、文书或外部环境；行人归期可结合应爻。",
    },
    "family": {
        "name": "家庭亲属",
        "yong_shen": "世爻",
        "notes": "须明确具体亲属关系；不可用性别替代代占关系。",
    },
    "competition": {
        "name": "竞争朋友",
        "yong_shen": "兄弟",
        "notes": "朋友、同辈、竞争、合伙分利。",
    },
    "children": {
        "name": "子女晚辈",
        "yong_shen": "子孙",
        "notes": "子女、晚辈、宠物、方案或解忧事项。",
    },
    "lost_item": {
        "name": "失物",
        "yong_shen": "妻财",
        "notes": "具体失物以妻财为初取，仍须按物品和语境核对。",
    },
    "legal": {
        "name": "争议纠纷",
        "yong_shen": "官鬼",
        "notes": "传统文化视角；不得替代法律意见。",
    },
}


ALIASES = {
    "academic": "study",
    "legal_risk": "legal",
    "lawsuit": "legal",
    "marriage": "relationship",
}


def resolve_matter(matter: str, *, perspective: str | None = None) -> dict:
    """返回事项路由事实。relationship 必须显式确认视角或高级用神。"""
    matter = ALIASES.get(matter, matter)
    if matter not in MATTERS:
        raise ValueError(
            f"unknown matter: {matter}; valid: {', '.join(sorted(MATTERS))}"
        )
    fact = MATTERS[matter].copy()
    if matter == "relationship":
        if perspective not in {"male", "female", "unspecified"}:
            raise ValueError(
                "relationship 需要 perspective=male/female/unspecified；"
                "unspecified 时应改传显式 yong_shen"
            )
        if perspective == "male":
            fact["yong_shen"] = "妻财"
        elif perspective == "female":
            fact["yong_shen"] = "官鬼"
        else:
            raise ValueError(
                "relationship perspective=unspecified 不能推断用神；请显式指定 yong_shen"
            )
    return fact
