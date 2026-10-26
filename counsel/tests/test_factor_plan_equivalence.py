"""真值表编译器 old-vs-new 等价测试。

旧解释器（_evaluate_truth_table）为 oracle；编译求值（_evaluate_compiled）
必须产生完全相同的 key 集合、值和插入顺序。
"""

from app.natal.tools.factor_tables import load_factor_rows, load_liunian_rows
from app.natal.tools.factors import _evaluate_compiled, _get_plan

from tests.factor_oracle import oracle_atomic as _atomic
from tests.factor_oracle import oracle_evaluate_truth_table as _evaluate_truth_table

BIRTH = ("1990-05-20T03:49:00Z", "Asia/Shanghai", "male", 116.4)


def _assert_equivalent(old: dict, new: dict, ctx: str) -> None:
    assert list(old.keys()) == list(new.keys()), (
        f"{ctx}: key insertion order mismatch\n"
        f"old={list(old.keys())[:10]}\nnew={list(new.keys())[:10]}"
    )
    for key in old:
        assert old[key] == new[key], (
            f"{ctx}: [{key}] old={old[key]!r} new={new[key]!r}"
        )


def _run_both(rows, atomic_fn, plan, gender, chart, ctx=None,
              current_year=0, factor_names=None, label=""):
    if factor_names is not None:
        filtered = [r for r in rows if r["因子"] in factor_names]
        old = _evaluate_truth_table(filtered, atomic_fn)
    else:
        old = _evaluate_truth_table(rows, atomic_fn)
    new = _evaluate_compiled(
        plan, gender, chart, ctx=ctx,
        current_year=current_year, factor_names=factor_names,
    )
    _assert_equivalent(old, new, label)


def _build_natal():
    from app.natal.tools.charts import build_natal_chart
    return build_natal_chart(*BIRTH)


def _build_liunian(natal_chart):
    from app.natal.tools.charts import _bazi_liunian
    return _bazi_liunian(natal_chart["chart"], 2030)


class TestNatalEquivalence:
    """本命 factors.csv 等价。"""

    def test_bazi_full(self):
        from app.natal.tools.factors import _factor_context_from_natal_chart
        natal = _build_natal()
        evaluation = _factor_context_from_natal_chart(natal)
        rows = [r for r in load_factor_rows() if r["术数"] == "bazi"]
        plan = _get_plan("natal", load_factor_rows(), "bazi")
        def atomic(expr):
            return _atomic(expr, "male", evaluation)
        _run_both(rows, atomic, plan, "male", evaluation, label="bazi_full")

    def test_ziwei_full(self):
        from app.natal.tools.factors import _factor_context_from_natal_chart
        natal = _build_natal()
        evaluation = _factor_context_from_natal_chart(natal)
        rows = [r for r in load_factor_rows() if r["术数"] == "ziwei"]
        plan = _get_plan("natal", load_factor_rows(), "ziwei")
        def atomic(expr):
            return _atomic(expr, "male", evaluation)
        _run_both(rows, atomic, plan, "male", evaluation, label="ziwei_full")

    def test_bazi_with_factor_names_closure(self):
        from app.natal.tools.factors import _factor_context_from_natal_chart
        natal = _build_natal()
        evaluation = _factor_context_from_natal_chart(natal)
        all_rows = [r for r in load_factor_rows() if r["术数"] == "bazi"]
        plan = _get_plan("natal", load_factor_rows(), "bazi")
        factor_names = {r["因子"] for r in all_rows[:50]}
        rows = [r for r in all_rows if r["因子"] in factor_names]
        def atomic(expr):
            return _atomic(expr, "male", evaluation)
        _run_both(rows, atomic, plan, "male", evaluation,
                  factor_names=factor_names, label="bazi_filtered_closure")

    def test_bazi_with_factor_names_missing_dep(self):
        from app.natal.tools.factors import _factor_context_from_natal_chart
        natal = _build_natal()
        evaluation = _factor_context_from_natal_chart(natal)
        all_rows = [r for r in load_factor_rows() if r["术数"] == "bazi"]
        plan = _get_plan("natal", load_factor_rows(), "bazi")
        factor_names = {
            r["因子"] for r in all_rows
            if r["conds"] and any("[" not in c for c in r["conds"])
        }
        factor_names = set(sorted(factor_names)[:30])
        rows = [r for r in all_rows if r["因子"] in factor_names]
        def atomic(expr):
            return _atomic(expr, "male", evaluation)
        _run_both(rows, atomic, plan, "male", evaluation,
                  factor_names=factor_names, label="bazi_missing_dep")


class TestLiunianEquivalence:
    """流年 factors_liunian.csv 等价。"""

    def test_bazi_liunian_full(self):
        natal = _build_natal()
        ln = _build_liunian(natal)
        ctx = {"liunian": ln, "zw_liunian": {}, "year": 2030,
               "chart": natal, "base": {}, "snapshot": {}}
        rows = [r for r in load_liunian_rows() if r["术数"] == "bazi"]
        plan = _get_plan("liunian", load_liunian_rows(), "bazi")
        def atomic(expr):
            return _atomic(expr, "male", natal, ctx)
        _run_both(rows, atomic, plan, "male", natal, ctx=ctx,
                  label="liunian_bazi_full")

    def test_ziwei_liunian_full(self):
        natal = _build_natal()
        from app.natal.tools.charts import _ziwei_liunian
        zw_ln = _ziwei_liunian(natal.get("ziwei_chart") or natal.get("ziwei") or {}, 2030)
        ctx = {"liunian": {}, "zw_liunian": zw_ln, "year": 2030,
               "chart": natal, "base": {}, "snapshot": {}}
        rows = [r for r in load_liunian_rows() if r["术数"] == "ziwei"]
        plan = _get_plan("liunian", load_liunian_rows(), "ziwei")
        def atomic(expr):
            return _atomic(expr, "male", natal, ctx)
        _run_both(rows, atomic, plan, "male", natal, ctx=ctx,
                  label="liunian_ziwei_full")


class TestPlanCache:
    """编译缓存正确性。"""

    def test_same_key_same_object(self):
        rows = load_factor_rows()
        plan_a = _get_plan("natal", rows, "bazi")
        plan_b = _get_plan("natal", rows, "bazi")
        assert plan_a is plan_b

    def test_different_key_different_object(self):
        rows = load_factor_rows()
        plan_a = _get_plan("natal", rows, "bazi")
        plan_b = _get_plan("natal", rows, "ziwei")
        assert plan_a is not plan_b
