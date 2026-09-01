"""因子真值表加载层：长表 CSV → evaluator 内部 rows。"""
from __future__ import annotations

import csv
import os
from collections import defaultdict
from pathlib import Path

from errors import FactorTableError

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
NATAL_PATH = os.path.join(TOOLS_DIR, "factors", "factors.csv")
LIUNIAN_PATH = os.path.join(TOOLS_DIR, "factors", "factors_liunian.csv")
META = {"factor_id", "shushi", "group_id", "term_index", "kind", "expression", "expected", "basis"}
LONG_FIELDS = [
    "factor_id", "shushi", "group_id", "term_index", "kind",
    "expression", "expected", "basis",
]

_CACHES: dict[Path, list[dict]] = {}


def _expected(value: str):
    value = (value or "").strip()
    try:
        return int(value)
    except ValueError:
        return value


def _ensure_acyclic(path: str, dependencies: dict[str, set[str]]) -> None:
    """校验因子引用图无环；环会使真值表结果依赖遍历顺序。"""
    state: dict[str, int] = {}
    stack: list[str] = []

    def visit(factor_id: str) -> None:
        mark = state.get(factor_id)
        if mark == 1:
            cycle_start = stack.index(factor_id)
            cycle = stack[cycle_start:] + [factor_id]
            raise FactorTableError(f"{path}: 因子引用成环: {' -> '.join(cycle)}")
        if mark == 2:
            return
        state[factor_id] = 1
        stack.append(factor_id)
        for dependency in dependencies.get(factor_id, ()):
            visit(dependency)
        stack.pop()
        state[factor_id] = 2

    for factor_id in dependencies:
        visit(factor_id)


def load_long_rows(path: str):
    """加载长表并按 (factor_id, group_id) 聚合为真值表行。

    OR 分支是独立 group；AND 条件是同 group 中的多个 term。
    direct group 只有一个表达式，且进入 evaluator 的 `直通` 字段。
    """
    cache_key = Path(path).resolve()
    if cache_key in _CACHES:
        return _CACHES[cache_key]
    raw_groups = defaultdict(list)
    factor_meta = {}
    factor_sides: dict[str, str] = {}
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
            expected = r.get("expected") or ""
            if kind == "direct" and expected.strip():
                raise FactorTableError(f"{path}: {factor_id} direct group 不能声明 expected")
            side = (r.get("shushi") or "bazi").strip()
            previous_side = factor_sides.setdefault(factor_id, side)
            if previous_side != side:
                raise FactorTableError(
                    f"{path}: {factor_id} 混用术数: {previous_side}/{side}"
                )
            raw_groups[(factor_id, group_id)].append((term_index, kind, expression, r.get("expected") or ""))
            factor_meta.setdefault((factor_id, group_id), (r.get("shushi") or "bazi", r.get("basis") or ""))

    dependencies: dict[str, set[str]] = defaultdict(set)
    for (factor_id, _group_id), terms in raw_groups.items():
        for _term_index, kind, expression, _row_expected in terms:
            if kind == "factor_ref":
                dependencies[factor_id].add(expression)
    missing = sorted(
        (factor_id, dependency)
        for factor_id, targets in dependencies.items()
        for dependency in targets
        if dependency not in factor_sides
    )
    if missing:
        factor_id, dependency = missing[0]
        raise FactorTableError(f"{path}: {factor_id} 引用不存在因子: {dependency}")
    _ensure_acyclic(path, dependencies)

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
            for term_index, _, expression, row_expected in terms:
                if term_index != expected_index:
                    raise FactorTableError(f"{path}: {factor_id} group {group_id} term_index 不连续")
                if expression in conds:
                    raise FactorTableError(f"{path}: {factor_id} group {group_id} 条件重复: {expression}")
                conds[expression] = _expected(row_expected)
                expected_index += 1
        rows.append({
            "因子": factor_id, "术数": shushi, "直通": direct,
            "conds": conds, "依据": basis,
        })
    # defaultdict/raw_groups 已保持首次出现顺序；不要隐式排序，保证断语/因子 diff 稳定。
    _CACHES[cache_key] = rows
    return rows


def load_factor_rows():
    return load_long_rows(NATAL_PATH)


def load_liunian_rows():
    return load_long_rows(LIUNIAN_PATH)
