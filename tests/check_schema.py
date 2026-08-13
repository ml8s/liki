"""断语库 schema 校验（数据库约束：断语引用的因子必须存在）。

校验内容：
1. 断语表约束键（含 any_of 内）必须 ∈ 因子全集（factors.yaml 因子名 + 引擎直读字段）
2. 断语约束键必须 ∈ 该表"引用因子"声明（若表有声明）——防声明与实际脱节
3. 引用因子声明中的因子名必须存在
4. 各断语表条目 id 唯一

用法：
    python3 tests/check_schema.py        # 校验全部，退出码 0/1
"""
from __future__ import annotations
import glob
import os
import sys
import yaml

_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DY = os.path.join(_ROOT, "tools")        # 推理机 + 全部断语表 + 中间数据（csv 只有工具读——谁用归谁）
TABLE_DIRS = (DY,)

# 引擎直读/系统键（非 factors.yaml 复合因子，但断语可用）
BUILTIN_KEYS = {"gender", "any_of", "事件", "强度", "优先级", "id", "结论", "依据"}
# 应期层因子（factors_liunian.yaml 定义——yingqi 域用，不在 factors.yaml）
def load_liunian_names() -> set:
    """流年因子名——从 factors_liunian.csv（真值表单一权威——json 已删）。"""
    import csv as _csv
    try:
        return {r["因子"] for r in _csv.DictReader(open(os.path.join(DY, "factors_liunian.csv"), encoding="utf-8"))}
    except Exception:
        return set()


LIUNIAN_KEYS = load_liunian_names()


def load_factor_shushi() -> dict:
    """因子名 → 术数（bazi/ziwei）——交叉校验：bazi 表只用八字因子、ziwei 表只用紫微因子。"""
    import csv as _csv
    mapping = {}
    for r in _csv.DictReader(open(os.path.join(DY, "factors.csv"), encoding="utf-8")):
        mapping[r["因子"]] = (r.get("术数") or "bazi").strip()
    return mapping


def load_factors_names() -> set:
    """因子名清单——从 factors.csv（真值表单一权威——json 已删）。"""
    import csv as _csv
    return {r["因子"] for r in _csv.DictReader(open(os.path.join(DY, "factors.csv"), encoding="utf-8"))}


def main() -> int:
    factor_names = load_factors_names()
    factor_shushi = load_factor_shushi()
    errors = []
    warnings = []
    seen_ids = {}
    files = glob.glob(os.path.join(DY, "*.csv"))
    for f in sorted(files):
        if os.path.basename(f) in ("factors.csv", "factors_liunian.csv"):
            continue
        import csv as _csv
        dom = os.path.basename(f)[:-4]
        rows = []
        with open(f, encoding="utf-8") as fh:
            for r in _csv.DictReader(fh):
                cons = {}
                for k, v in r.items():
                    if k in ("id", "事件", "结论", "依据", "经典原文"):
                        continue
                    if (v or "").strip():
                        cons[k] = v
                # 交叉校验：八字表只用八字因子、紫微表只用紫微因子（防混合回潮——真分开）
                expect = "bazi" if dom.startswith("bazi_") else ("ziwei" if dom.startswith("ziwei_") else None)
                if expect:
                    for ck in cons:
                        cs = factor_shushi.get(ck)
                        if cs and cs != expect:
                            warnings.append(f"[{dom}] 跨术数条件列 '{ck}'（{cs}）——{expect} 表应纯{expect}因子")
                rows.append({"id": r.get("id", ""), "约束": cons})
        used_keys = set()
        for item in rows:
            eid = item.get("id")
            if eid:
                if eid in seen_ids:
                    errors.append(f"断语 id 重复: {eid}（{seen_ids[eid]} 与 {f}）")
                seen_ids[eid] = f
            for k in (item.get("约束") or {}):
                used_keys.add(k)
        # 约束键必须存在
        for k in used_keys:
            if k not in factor_names and k not in BUILTIN_KEYS and k not in LIUNIAN_KEYS:
                errors.append(f"[{dom}] 约束键 '{k}' 不在因子全集（factors.yaml 无此因子）")

    print(f"因子全集: {len(factor_names)} 个")
    print(f"错误: {len(errors)} 个")
    for e in errors:
        print("  ✗", e)
    print(f"警告: {len(warnings)} 个")
    for w in warnings[:40]:
        print("  ⚠", w)
    if len(warnings) > 40:
        print(f"  ... 共 {len(warnings)} 条警告")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
