"""六爻报告硬事实审计；只校验盘面事实，不裁决吉凶与应期。"""
from __future__ import annotations

import re

from liuyao_report import validate_report


POSITIONS = {"初": 1, "二": 2, "三": 3, "四": 4, "五": 5, "上": 6}
RELATIVES = ("父母", "兄弟", "子孙", "妻财", "官鬼")


def audit_report(report: str | dict, chart: dict, snapshot: dict | None = None) -> dict:
    """校验报告中的显式确定性事实；推断类内容保持未检查。"""
    if not isinstance(chart, dict):
        raise ValueError("chart must be an object")
    errors: list[dict] = []

    if isinstance(report, dict):
        snapshot = snapshot or chart.get("snapshot")
        report_contract_audit = validate_report(report, snapshot or {})
        errors.extend({
            "type": "report_contract",
            **item,
        } for item in report_contract_audit.get("errors", []))
    elif isinstance(report, str):
        report_contract_audit = None
        if not report.strip():
            raise ValueError("report must be non-empty text")
    else:
        raise ValueError("report must be text or structured report object")

    raw = chart.get("chart", chart)
    if not isinstance(raw, dict):
        raise ValueError("chart.chart must be an object")
    lines = raw.get("lines")
    if not isinstance(lines, list) or len(lines) != 6:
        raise ValueError("chart must contain exactly 6 lines")
    report_text = report if isinstance(report, str) else "\n".join(
        str(report.get(field, ""))
        for field in ("headline", "verdict", "action")
    )

    for label, actual in (("本卦", raw.get("ben_gua")), ("变卦", raw.get("bian_gua"))):
        if not actual:
            continue
        for match in re.finditer(rf"{label}(?:为|是|：|:)\s*([^，。；\s]+)", report_text):
            claimed = match.group(1)
            if claimed != actual:
                errors.append({"type": f"{label}_name", "claimed": claimed, "actual": actual})

    for label, field in (("世", "世"), ("应", "应")):
        actual = next(
            (line.get("position") for line in lines if line.get("shi_ying") == field),
            None,
        )
        for match in re.finditer(rf"{label}爻(?:在|居|临)([初二三四五上])爻", report_text):
            claimed = POSITIONS[match.group(1)]
            if claimed != actual:
                errors.append({"type": f"{label}_position", "claimed": claimed, "actual": actual})

    for match in re.finditer(rf"([初二三四五上])爻(?:为|是|临)({'|'.join(RELATIVES)})", report_text):
        position = POSITIONS[match.group(1)]
        claimed = match.group(2)
        actual = lines[position - 1].get("liu_qin")
        if claimed != actual:
            errors.append({
                "type": "line_relative",
                "position": position,
                "claimed": claimed,
                "actual": actual,
            })

    return {
        "schema_version": "liuyao-report-audit-v1",
        "accepted": not errors,
        "errors": errors,
        "scope": "explicit deterministic chart claims only",
        "unchecked": ["吉凶判断", "应期推断", "传统取象", "现实建议"],
        "report_contract_audit": report_contract_audit,
    }
