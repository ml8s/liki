"""pan 契约层 — 只接受 full_paipan 的完整返回结构。"""
from __future__ import annotations

from errors import PanSchemaError

REQUIRED_FIELDS = ("solar", "lunar", "chart", "full", "yongshen", "ziwei", "gender")
PILLARS = ("nian", "yue", "ri", "shi")
GENDERS = ("male", "female")


def validate_natal_pan(pan: object, action: str = "pan") -> None:
    """拒绝快照、裁剪盘和半截盘；完整盘契约由本层统一维护。"""
    if not isinstance(pan, dict):
        raise PanSchemaError(
            f"{action} pan 必须是 full_paipan 返回的完整本命盘 object。"
            "流年查询请走 yearly_range。"
        )
    missing = [key for key in REQUIRED_FIELDS if key not in pan]
    if missing:
        raise PanSchemaError(
            f"{action} pan 必须是 full_paipan 返回的完整本命盘，缺少字段: "
            f"{', '.join(missing)}。禁止传快照或裁剪盘。"
        )
    wrong = [
        key for key in REQUIRED_FIELDS
        if key not in ("solar", "gender") and not isinstance(pan[key], dict)
    ]
    if wrong:
        raise PanSchemaError(
            f"{action} pan 字段类型不完整: {', '.join(wrong)} 必须是 object。"
        )
    if not isinstance(pan["solar"], str) or not pan["solar"]:
        raise PanSchemaError(f"{action} pan.solar 必须是非空字符串。")
    if pan["gender"] not in GENDERS:
        raise PanSchemaError(
            f"{action} pan.gender 必须是 male/female，收到: {pan['gender']}"
        )

    chart, full = pan["chart"], pan["full"]
    missing_pillars = [
        key for key in PILLARS
        if not isinstance(chart.get(key), dict)
        or not isinstance(full.get(key), dict)
        or not full[key].get("gan")
        or not full[key].get("zhi")
    ]
    if missing_pillars:
        raise PanSchemaError(
            f"{action} pan 四柱结构不完整，缺少: {', '.join(missing_pillars)}。"
            "请重新执行 full_paipan，不要手工拼装半截盘。"
        )
    if not isinstance(chart.get("da_yun"), dict):
        raise PanSchemaError(f"{action} pan.chart.da_yun 缺失或不是 object。")
    if not isinstance(pan["ziwei"].get("gong_wei"), list):
        raise PanSchemaError(f"{action} pan.ziwei.gong_wei 缺失或不是 array。")
