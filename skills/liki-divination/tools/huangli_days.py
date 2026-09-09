"""黄历择日编排：安全分流、日期范围、事件适配与候选日输出。"""
from __future__ import annotations

from datetime import date, timedelta

from divination_rpc import engine_data
from divination_safety import assess, blocked_payload


SCHEMA_VERSION = "huangli-read-v1"
MAX_DAYS = 30

EVENT_RULES = {
    "wedding": {"label": "嫁娶", "suitable": {"成", "开"}, "forbidden": {"破", "建", "除", "满", "收"}},
    "engage": {"label": "领证", "suitable": {"定", "成"}, "forbidden": {"破", "满"}},
    "opening": {"label": "开业", "suitable": {"成", "开"}, "forbidden": {"破", "除", "危", "闭"}},
    "sign": {"label": "签约", "suitable": {"定", "成"}, "forbidden": {"破", "满", "闭"}},
    "move": {"label": "搬家", "suitable": {"成", "平"}, "forbidden": {"破", "执"}},
    "travel": {"label": "出行", "suitable": {"建", "开"}, "forbidden": {"破", "平", "危", "收"}},
    "build": {"label": "动土", "suitable": {"平"}, "forbidden": {"破", "定", "开"}},
    "exam": {"label": "考试", "suitable": {"定", "成"}, "forbidden": {"破"}},
    "medical": {"label": "就医", "suitable": {"除"}, "forbidden": {"破"}},
    "funeral": {"label": "丧葬", "suitable": {"闭"}, "forbidden": {"破", "开"}},
}

EVENT_ALIASES = {
    "结婚": "wedding", "婚礼": "wedding", "嫁娶": "wedding", "领证": "engage",
    "开业": "opening", "开市": "opening", "签约": "sign", "合同": "sign",
    "搬家": "move", "移徙": "move", "乔迁": "move", "出行": "travel",
    "旅游": "travel", "出差": "travel", "动土": "build", "考试": "exam",
    "就医": "medical", "丧葬": "funeral",
}


def _server_date() -> date:
    response = engine_data("time.now", {})
    cst = response.get("cst")
    if not isinstance(cst, str) or len(cst) < 10:
        raise ValueError("time.now returned no usable cst date")
    return date.fromisoformat(cst[:10])


def _date(value: str, field: str) -> date:
    try:
        return date.fromisoformat(value)
    except (TypeError, ValueError) as error:
        raise ValueError(f"{field} must be YYYY-MM-DD") from error


def _normalize_event(event: str | None) -> tuple[str | None, str | None]:
    if event is None or event == "":
        return None, None
    key = EVENT_ALIASES.get(event, event)
    if key not in EVENT_RULES:
        raise ValueError(f"unknown event: {event}")
    return key, EVENT_RULES[key]["label"]


def _classify(day: dict, rule: dict | None) -> dict:
    jianchu = day.get("jian_chu")
    suitability = "unspecified"
    reason = "未指定事项，仅列黄历事实。"
    if rule:
        if jianchu in rule["suitable"]:
            suitability = "recommended"
            reason = f"建除「{jianchu}」适合{rule['label']}。"
        elif jianchu in rule["forbidden"]:
            suitability = "unsuitable"
            reason = f"建除「{jianchu}」忌{rule['label']}。"
        else:
            suitability = "possible"
            reason = f"建除「{jianchu}」无明确适配或冲突。"
    warnings = [day.get(field) for field in ("gan_ji", "zhi_ji") if day.get(field)]
    return {
        "date": day.get("date"),
        "jian_chu": jianchu,
        "suitability": suitability,
        "reason": reason,
        "warnings": warnings,
        "day": day,
    }


def days(
    *,
    question: str,
    event: str | None = None,
    start_date: str | None = None,
    end_date: str | None = None,
    days: int | None = None,
) -> dict:
    """查询黄历日期范围并按事项做确定性建除标注。"""
    safety = assess(question)
    if safety["status"] != "allow":
        return blocked_payload(SCHEMA_VERSION, question, "huangli")

    event_key, event_label = _normalize_event(event)
    today = _server_date()
    start = _date(start_date, "start_date") if start_date is not None else today + timedelta(days=1)
    if end_date is not None:
        end = _date(end_date, "end_date")
        if end < start:
            raise ValueError("end_date must not be before start_date")
        count = (end - start).days + 1
    else:
        count = days if days is not None else 30
    if not isinstance(count, int) or not 1 <= count <= MAX_DAYS:
        raise ValueError(f"date range must contain 1-{MAX_DAYS} days")

    raw_days = engine_data("huangli.days", {"start_date": start.isoformat(), "count": count})
    if not isinstance(raw_days, list):
        raise ValueError("huangli.days returned a non-array")

    rule = EVENT_RULES[event_key] if event_key else None
    candidates = [_classify(day, rule) for day in raw_days if isinstance(day, dict)]
    recommended = [item for item in candidates if item["suitability"] == "recommended"]
    unsuitable = [item for item in candidates if item["suitability"] == "unsuitable"]
    return {
        "schema_version": SCHEMA_VERSION,
        "question": question,
        "event": event_key,
        "event_label": event_label,
        "range": {
            "start_date": start.isoformat(),
            "end_date": (start + timedelta(days=count - 1)).isoformat(),
            "days": count,
        },
        "safety": safety,
        "recommended": recommended,
        "unsuitable": unsuitable,
        "candidates": candidates,
        "policy": {
            "candidate_scope": "traditional_huangli_labels_only",
            "no_bazi_composite_without_birth_data": True,
            "no_guarantee": True,
        },
    }
