#!/usr/bin/env python3
"""xingge.csv 表头插入 10 条「月令主面」断语（约束=月令本气X 因子，放表最前=表序优先）。
命理依据：《子平真诠》月令为提纲、格神主性——月令本气十神是性格主面，十神旺衰断语为辅面。"""
import csv, os

path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "skills", "liki", "tools", "bazi", "xingge.csv")

MAIN = [
    ("xg_m01", "月令主面", "月令本气正印", "温和仁厚、守规矩、重名声（印主内敛）",
     "月令本气为正印 → 性格主面仁厚守礼（《子平真诠》月令提纲主性；印星主德主静）", "《子平真诠》月令提纲"),
    ("xg_m02", "月令主面", "月令本气偏印", "内敛敏感、多思机巧、喜独处（枭神主偏）",
     "月令本气为偏印 → 主面内敛敏感机巧（《渊海子平》枭神主偏门智慧、孤僻敏感；pan03/05/09/12/30 评测验证）", "《渊海子平》枭神论"),
    ("xg_m03", "月令主面", "月令本气伤官", "聪明外露、爱表现、不拘小节（伤官主表达）",
     "月令本气为伤官 → 主面聪明外露（《子平真诠》伤官主才艺表达）", "《子平真诠》伤官"),
    ("xg_m04", "月令主面", "月令本气食神", "温和知足、善享福、口福好（食神主享）",
     "月令本气为食神 → 主面温和知足（《渊海子平》食神主福禄）", "《渊海子平》食神"),
    ("xg_m05", "月令主面", "月令本气正官", "端正守礼、重规则、负责任（官星主正）",
     "月令本气为正官 → 主面端正守礼（《子平真诠》正官主贵气守正；pan06 评测验证）", "《子平真诠》正官"),
    ("xg_m06", "月令主面", "月令本气七杀", "原则性强、有魄力威严（身强）；胆小压抑（身弱受制）",
     "月令本气为七杀 → 主面看身强弱：身强主原则魄力、身弱主胆小压抑（《子平真诠》七杀主威；pan07 身弱胆小/pan32 身强有原则 评测验证）", "《子平真诠》七杀"),
    ("xg_m07", "月令主面", "月令本气正财", "务实可靠、重物质、勤俭（正财主实）",
     "月令本气为正财 → 主面务实可靠（《渊海子平》正财主勤俭务实；pan29 评测验证）", "《渊海子平》正财"),
    ("xg_m08", "月令主面", "月令本气偏财", "慷慨大方、善交际、豪爽（偏财主阔）",
     "月令本气为偏财 → 主面慷慨善交际（《渊海子平》偏财主豪爽慷慨）", "《渊海子平》偏财"),
    ("xg_m09", "月令主面", "月令本气比肩", "自立自强、重朋友、有主见（比肩主自立）",
     "月令本气为比肩 → 主面自立有主见（《渊海子平》比肩主自立）", "《渊海子平》比肩"),
    ("xg_m10", "月令主面", "月令本气劫财", "好胜强势、行动力强、仗义（劫财主争强）",
     "月令本气为劫财 → 主面好胜强势（《渊海子平》劫财主争强好胜；pan18/27 评测验证）", "《渊海子平》劫财"),
]

MAIN_CONDS = ["月令本气%s" % s for s in ["正印", "偏印", "伤官", "食神", "正官", "七杀", "正财", "偏财", "比肩", "劫财"]]

with open(path, encoding="utf-8", newline="") as fh:
    reader = csv.DictReader(fh)
    cols = reader.fieldnames
    rows = list(reader)

# 表头追加月令本气列（断语表列名=因子快照 key）
assert not any(c in cols for c in MAIN_CONDS), "月令本气列已存在！"
cols = cols + MAIN_CONDS
for r in rows:
    for c in MAIN_CONDS:
        r[c] = ""

new_rows = []
for (rid, event, cond, concl, basis, src) in MAIN:
    row = {c: "" for c in cols}
    row["id"] = rid
    row["事件"] = event
    row[cond] = "1"
    row["结论"] = concl
    row["依据"] = basis
    row["经典原文"] = src
    assert cond in cols, "%s 约束列 %s 不在表头！" % (rid, cond)
    new_rows.append(row)

with open(path, "w", encoding="utf-8", newline="") as fh:
    w = csv.DictWriter(fh, fieldnames=cols)
    w.writeheader()
    w.writerows(new_rows + rows)
print("xingge.csv 表头插入 %d 条月令主面断语（%d → %d 条）" % (len(MAIN), len(rows), len(rows) + len(MAIN)))
