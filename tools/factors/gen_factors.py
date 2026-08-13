"""窄表 ↔ 宽表双向生成（命理师维护 factors_narrow.csv，宽表 factors.csv 为执行真值表）。

窄表格式：因子,术数,组,原子,期望值
  - 组内多行 = 且（同一组所有条件全满足）；不同组 = 或（宽表多行）
  - "原子"字段四类：
      1. 原子算子（含 `[]`，如 `得地[配偶星]`）→ 宽表原子列
      2. 引用因子名（不含 `[]`，如 `伤官克官`）→ 宽表引用列（引用其他因子的值）
      3. `直通:<算子>` → 宽表原语直通列
      4. `结论` / `依据` → 宽表结论/依据列（元数据）
  - 期望值为字符串（如"冲"）时宽表列值=该字符串

用法：
  python3 gen_factors.py            # 窄表 → 宽表（默认）
  python3 gen_factors.py --reverse  # 宽表 → 窄表（宽表改动后重建窄表）
"""
import csv, os, sys
from collections import OrderedDict

here = os.path.dirname(os.path.abspath(__file__))
NARROW = os.path.join(here, "factors_narrow.csv")
WIDE = os.path.join(here, "factors.csv")
META = ("因子", "术数", "原语直通", "结论", "依据")


def narrow_to_wide():
    narrow = list(csv.DictReader(open(NARROW, encoding="utf-8")))
    old = list(csv.DictReader(open(WIDE, encoding="utf-8")))
    all_cols = [c for c in old[0].keys() if c not in ("因子", "术数", "原语直通")]
    # 收集窄表出现但宽表没有的新原子列（如新增的"宫含[财帛宫,紫微,任意]"）——否则新因子判据丢失
    for r in narrow:
        atom = r["原子"]
        if atom and atom not in ("结论", "依据") and not atom.startswith("直通:") and atom not in all_cols:
            all_cols.append(atom)
    groups = OrderedDict()   # (因子, 组) -> {列: 值}
    zhitong = OrderedDict()  # 因子 -> 直通算子
    shushi_map = OrderedDict()
    for r in narrow:
        f = r["因子"]
        shushi_map[f] = r["术数"]
        atom = r["原子"]
        if atom.startswith("直通:"):
            zhitong[f] = atom[3:]
            continue
        groups.setdefault((f, r["组"]), {})[atom] = r["期望值"]
    rows = []
    for (f, g), conds in groups.items():
        row = {"因子": f, "术数": shushi_map.get(f, "bazi"), "原语直通": zhitong.get(f, "")}
        for c in all_cols:
            row[c] = conds.get(c, "")
        rows.append(row)
    # 直通因子（无组的）也输出一行
    for f, zt in zhitong.items():
        if f not in {g[0] for g in groups}:
            row = {"因子": f, "术数": shushi_map.get(f, "bazi"), "原语直通": zt}
            for c in all_cols:
                row[c] = ""
            rows.append(row)
    with open(WIDE, "w", encoding="utf-8", newline="") as w:
        wr = csv.DictWriter(w, fieldnames=["因子", "术数", "原语直通"] + all_cols)
        wr.writeheader(); wr.writerows(rows)
    print(f"窄表 → 宽表：{len(rows)} 行")


def wide_to_narrow():
    rows = list(csv.DictReader(open(WIDE, encoding="utf-8")))
    narrow = []
    seq = {}   # 因子 -> 已用组号计数（同因子多行=或组，序号 1/2/3）
    for r in rows:
        f = r["因子"]
        shushi = r["术数"]
        zt = (r.get("原语直通") or "").strip()
        if zt:
            narrow.append([f, shushi, "", f"直通:{zt}", ""])
            continue
        seq[f] = seq.get(f, 0) + 1
        g = str(seq[f])
        # 原子列 + 引用列（含[]=原子，否则=引用因子名——除 meta 外）
        for c in r:
            if c in ("因子", "术数", "原语直通"):
                continue
            v = (r[c] or "").strip()
            if v:
                narrow.append([f, shushi, g, c, v])
    with open(NARROW, "w", encoding="utf-8", newline="") as w:
        wr = csv.writer(w)
        wr.writerow(["因子", "术数", "组", "原子", "期望值"])
        wr.writerows(narrow)
    print(f"宽表 → 窄表：{len(narrow)} 行")


if __name__ == "__main__":
    if "--reverse" in sys.argv:
        wide_to_narrow()
    else:
        narrow_to_wide()
