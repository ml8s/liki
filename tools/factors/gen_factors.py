"""窄表 → 宽表生成（命理师维护 factors_narrow.csv，本脚本生成 factors.csv 真值表）。

窄表格式：因子,术数,组,原子,期望值
  组内多行 = 且（同一组的所有原子全满足）；不同组 = 或（宽表多行）
  直通行："直通:<算子>"（原语直通列）；期望值为字符串（如"冲"）时宽表列值=该字符串
"""
import csv, os
from collections import OrderedDict

here = os.path.dirname(os.path.abspath(__file__))
narrow = list(csv.DictReader(open(os.path.join(here, "factors_narrow.csv"), encoding="utf-8")))

# 收集所有原子列（保持原宽表顺序——用现有 factors.csv 的列序）
old = list(csv.DictReader(open(os.path.join(here, "factors.csv"), encoding="utf-8")))
all_cols = [c for c in old[0].keys() if c not in ("因子", "术数", "原语直通")]

# 按因子+组聚合
groups = OrderedDict()   # (因子, 组) -> {原子列: 值}
zhitong = OrderedDict()  # 因子 -> 直通算子
shushi_map = OrderedDict()
for r in narrow:
    f = r["因子"]; shushi_map[f] = r["术数"]
    if r["原子"].startswith("直通:"):
        zhitong[f] = r["原子"][3:]
        continue
    key = (f, r["组"])
    groups.setdefault(key, {})[r["原子"]] = r["期望值"]

rows = []
for (f, g), conds in groups.items():
    row = {"因子": f, "术数": shushi_map.get(f, "bazi"), "原语直通": zhitong.get(f, "")}
    for c in all_cols:
        row[c] = conds.get(c, "")
    rows.append(row)

with open(os.path.join(here, "factors.csv"), "w", encoding="utf-8", newline="") as w:
    wr = csv.DictWriter(w, fieldnames=["因子", "术数", "原语直通"] + all_cols)
    wr.writeheader(); wr.writerows(rows)
print(f"宽表生成: {len(rows)} 行")
