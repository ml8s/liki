"""Projection helpers for natal assertion results."""

_BRIEF_FIELDS = ("id", "领域", "事件类型", "时间层", "事件", "结论")


def current_step(steps, year: int):
    """当前大运/大限步骤：按公历年段匹配；无年份或不在任何步骤内返回 None。"""
    if not year or not isinstance(steps, list):
        return None
    for index, step in enumerate(steps):
        if not isinstance(step, dict):
            continue
        start, end = step.get("start_year"), step.get("end_year")
        if isinstance(start, int) and isinstance(end, int) and start <= year <= end:
            item = {
                "index": index,
                "name": step.get("name", ""),
                "start_year": start,
                "end_year": end,
                "next_step_in_years": end - year,
            }
            if step.get("gong"):
                item["gong"] = step["gong"]
            return item
    return None


def brief(conclusions: list) -> list:
    return [{key: row[key] for key in _BRIEF_FIELDS if key in row} for row in conclusions if isinstance(row, dict)]
