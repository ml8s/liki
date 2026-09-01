"""因子真值表加载层：长表 CSV → evaluator 内部 rows。"""
from __future__ import annotations

import csv
import os
from collections import defaultdict

from errors import FactorTableError

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
NATAL_PATH = os.path.join(TOOLS_DIR, "factors", "factors.csv")
LIUNIAN_PATH = os.path.join(TOOLS_DIR, "factors", "factors_liunian.csv")
META = {"factor_id", "shushi", "group_id", "term_index", "kind", "expression", "expected", "basis"}
LONG_FIELDS = [
    "factor_id", "shushi", "group_id", "term_index", "kind",
    "expression", "expected", "basis",
]

_CACHES = {}


def _expected(value: str):
    value = (value or "").strip()
    try:
        return int(value)
    except ValueError:
        return value


def load_long_rows(path: str, cache_key: str):
    """加载长表并按 (factor_id, group_id) 聚合为真值表行。

    OR 分支是独立 group；AND 条件是同 group 中的多个 term。
    direct group 只有一个表达式，且进入 evaluator 的 `直通` 字段。
    """
    if cache_key in _CACHES:
        return _CACHES[cache_key]
    raw_groups = defaultdict(list)
    factor_meta = {}
    with open(path, encoding="utf-8-sig", newline="") as fh:
        for r in csv.DictReader(fh):
            factor_id = (r.get("factor_id") or "").strip()
            if not factor_id:
                raise FactorTableError(f"{path}: factor_id 不能为空")
            try:
                group_id = int(r["group_id"])
                term_index = int(r["term_index"])
            except (KeyError, TypeError, ValueError) as e:
                raise FactorTableError(f"{path}: {factor_id} group/term 编号无效") from e
            kind = (r.get("kind") or "").strip()
            expression = (r.get("expression") or "").strip()
            if kind not in ("direct", "condition", "factor_ref"):
                raise FactorTableError(f"{path}: {factor_id} kind 无效: {kind}")
            if not expression:
                raise FactorTableError(f"{path}: {factor_id} expression 不能为空")
            raw_groups[(factor_id, group_id)].append((term_index, kind, expression, r.get("expected") or ""))
            factor_meta.setdefault((factor_id, group_id), (r.get("shushi") or "bazi", r.get("basis") or ""))

    rows = []
    for (factor_id, group_id), terms in raw_groups.items():
        terms.sort(key=lambda x: x[0])
        shushi, basis = factor_meta[(factor_id, group_id)]
        kinds = {kind for _, kind, _, _ in terms}
        if "direct" in kinds and kinds - {"direct"}:
            raise FactorTableError(f"{path}: {factor_id} group {group_id} 混用 direct/condition")
        if kinds == {"condition", "factor_ref"}:
            # factor 引用与算子条件可在同一 AND 行中组合。
            kind = "condition"
        else:
            if len(kinds) != 1:
                raise FactorTableError(f"{path}: {factor_id} group {group_id} 混用 direct/condition")
            kind = next(iter(kinds))
        conds = {}
        direct = ""
        if kind == "direct":
            if len(terms) != 1:
                raise FactorTableError(f"{path}: {factor_id} direct group 必须只有一个 term")
            direct = terms[0][2]
        else:
            expected_index = 1
            for term_index, _, expression, expected in terms:
                if term_index != expected_index:
                    raise FactorTableError(f"{path}: {factor_id} group {group_id} term_index 不连续")
                if expression in conds:
                    raise FactorTableError(f"{path}: {factor_id} group {group_id} 条件重复: {expression}")
                conds[expression] = _expected(expected)
                expected_index += 1
        rows.append({
            "因子": factor_id, "术数": shushi, "直通": direct,
            "conds": conds, "依据": basis,
        })
    # defaultdict/raw_groups 已保持首次出现顺序；不要隐式排序，保证断语/因子 diff 稳定。
    _CACHES[cache_key] = rows
    return rows


def load_factor_rows():
    return load_long_rows(NATAL_PATH, "natal")


def load_liunian_rows():
    return load_long_rows(LIUNIAN_PATH, "liunian")

