#!/usr/bin/env python3
"""factors.csv 追加 10 个「月令本气十神」因子（性格主面上下文因子）。
安全修改：表头追加 10 列，现有行补空，追加 10 行新因子。"""
import csv, os

path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "tools", "factors", "factors.csv")
NEW = ["正印", "偏印", "伤官", "食神", "正官", "七杀", "正财", "偏财", "比肩", "劫财"]

with open(path, encoding="utf-8", newline="") as fh:
    reader = csv.DictReader(fh)
    cols = reader.fieldnames
    rows = list(reader)

new_cols = ["月令本气[%s]" % s for s in NEW]
assert not any(c in cols for c in new_cols), "新列已存在！"
cols = cols + new_cols

for r in rows:
    for c in new_cols:
        r[c] = ""

for s in NEW:
    row = {c: "" for c in cols}
    row["因子"] = "月令本气%s" % s
    row["术数"] = "bazi"
    row["月令本气[%s]" % s] = "1"
    row["结论"] = "月令本气十神=%s（格神主性——《子平真诠》月令为提纲）" % s
    row["依据"] = "月令本气（月支藏干主气）的十神——性格主面/格局本气"
    rows.append(row)

with open(path, "w", encoding="utf-8", newline="") as fh:
    w = csv.DictWriter(fh, fieldnames=cols)
    w.writeheader()
    w.writerows(rows)
print("factors.csv 已追加 %d 个「月令本气」因子（%d 列 → %d 列）" % (len(NEW), len(cols) - len(NEW), len(cols)))
