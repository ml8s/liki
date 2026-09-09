"""问卦 / 择日统一安全边界。"""
from __future__ import annotations

import re


RULES = [
    ("self_harm", r"自杀|自残|不想活|伤害自己|轻生"),
    ("emergency_medical", r"大出血|呼吸困难|昏迷|失去意识|剧烈腹痛|急救|救护车"),
    ("pregnancy", r"怀孕|胎儿|保胎|流产|预产期|胎儿性别"),
    ("serious_medical", r"癌症|重病|手术|能活多久|生死|确诊.*怎么治|停药|换药"),
    ("missing_person_safety", r"被绑架|凶手|还活着吗|下落不明.*生死"),
    ("major_financial_risk", r"全部身家|倾家荡产|高利贷|梭哈|借.*全部.*投"),
    ("criminal_risk", r"犯罪|逃税|杀人|伤害他人|毒品"),
]

GUIDANCE = {
    "self_harm": "请立即联系当地紧急援助、心理危机干预热线或可信赖的人陪伴。",
    "emergency_medical": "请立即联系当地急救或医疗机构。",
    "pregnancy": "请咨询产科或医疗机构。",
    "serious_medical": "请咨询主治医疗机构，不要用卦象判断病情或调整治疗。",
    "missing_person_safety": "请优先联系警方或救援机构。",
    "major_financial_risk": "请咨询持牌财务或法律专业人士，先止损再做决定。",
    "criminal_risk": "请咨询律师或联系有权处理机关。",
}


def assess(question: str) -> dict:
    """返回 allow/redirect 状态；调用方必须在排盘前检查。"""
    if not isinstance(question, str):
        raise ValueError("question must be a string")
    for category, pattern in RULES:
        if re.search(pattern, question, re.IGNORECASE):
            return {
                "status": "redirect",
                "category": category,
                "message": "该问题涉及高风险现实事项；不用卦象判断结果，先寻求专业或紧急支持。",
                "guidance": GUIDANCE[category],
            }
    return {"status": "allow", "category": None, "message": "", "guidance": ""}


def blocked_payload(schema_version: str, question: str, method: str = "divination") -> dict:
    safety = assess(question)
    return {
        "schema_version": schema_version,
        "blocked": safety["status"] == "redirect",
        "route": "blocked",
        "method": method,
        "question": question,
        "safety": safety,
        "policy": {
            "no_chart": True,
            "no_timing": True,
            "no_guilty_or_fatal_prediction": True,
        },
    }
