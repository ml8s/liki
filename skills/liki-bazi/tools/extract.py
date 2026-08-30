"""提取层 — pan → 因子求值上下文（fac）。

只做引擎数据到求值上下文的机械翻译：
- 十神状态
- 五行计数与旺衰
- 三派用神
- 日干
- 大运公历年段

不生成因子、不做命理组合判断。
"""
from __future__ import annotations

from dataclasses import asdict, dataclass
from typing import Optional

_DE_LING_STATES = {"旺", "相"}


@dataclass
class ShishenState:
    """某十神在原局的聚合状态。"""
    name: str
    wuxing: str
    tou_gan: bool
    cang_zhi: bool
    has_root: bool
    de_ling: bool
    count: int = 0


def _wuxing_of_gan(gan: str) -> str:
    return {"甲": "木", "乙": "木", "丙": "火", "丁": "火", "戊": "土",
            "己": "土", "庚": "金", "辛": "金", "壬": "水", "癸": "水"}.get(gan, "")


def _all_shishens(full: dict) -> list[dict]:
    out = []
    for pillar in ("nian", "yue", "ri", "shi"):
        for item in (full.get(pillar, {}) or {}).get("shi_shens", []) or []:
            out.append({"pillar": pillar, **item})
    return out


def extract_shishen(full: dict, wang_shuai: Optional[dict] = None) -> dict[str, ShishenState]:
    """聚合四柱十神：透干、藏支、有根、得令、计数。"""
    states: dict[str, ShishenState] = {}
    roots: set[str] = set()
    for item in _all_shishens(full):
        name = item.get("shi_shen", "")
        if not name:
            continue
        state = states.setdefault(
            name,
            ShishenState(name=name, wuxing="", tou_gan=False, cang_zhi=False,
                         has_root=False, de_ling=False, count=0),
        )
        gan = item.get("gan", "")
        if not state.wuxing and gan:
            state.wuxing = _wuxing_of_gan(gan)
        source = item.get("source", "")
        if source == "gan":
            state.tou_gan = True
            state.count += 1
        elif source.endswith("qi"):
            state.cang_zhi = True
            state.count += 1

    for pillar in ("nian", "yue", "ri", "shi"):
        hidden = (full.get(pillar, {}) or {}).get("cang_gan", {}) or {}
        for value in (hidden.get("main"), hidden.get("mid"), hidden.get("minor")):
            if value:
                roots.add(_wuxing_of_gan(value))

    for state in states.values():
        if not state.wuxing:
            continue
        state.has_root = state.wuxing in roots
        state.de_ling = bool(wang_shuai and wang_shuai.get(state.wuxing) in _DE_LING_STATES)
    return states


def extract(pan: dict) -> dict:
    """从 full_paipan 返回的 pan 中提取因子求值上下文。"""
    full = pan.get("full", {}) or {}
    yongshen = pan.get("yongshen", {}) or {}
    fu_yi = yongshen.get("fu_yi", {}) or {}
    steps = ((pan.get("chart", {}) or {}).get("da_yun", {}) or {}).get("steps", []) or []
    return {
        "shishen": {name: asdict(state)
                    for name, state in extract_shishen(full, fu_yi.get("wang_shuai", {})).items()},
        "wuxing": {
            "count": fu_yi.get("wuxing_count", {}) or {},
            "wang_shuai": fu_yi.get("wang_shuai", {}) or {},
        },
        "yongshen": yongshen,
        "ri_gan": (full.get("ri", {}) or {}).get("gan", ""),
        "palace_ri": {
            "zhi": (full.get("ri", {}) or {}).get("zhi", ""),
        },
        "dayun_steps": [
            {
                "name": step.get("name", ""),
                "start_year": step.get("start_year", 0),
                "end_year": step.get("end_year", 0),
                "shi_shen": step.get("shi_shen", ""),
            }
            for step in steps
        ],
    }
