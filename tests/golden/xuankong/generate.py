#!/usr/bin/env python3
"""Generate a complete lower-plate XuanKong flying-star matrix fixture."""
from __future__ import annotations

import json
from pathlib import Path

ROOT=Path(__file__).resolve().parents[3]
MOUNTAINS=json.loads((ROOT/"tests/fixtures/domain_oracle/fengshui_core.json").read_text())["mountains_24"]
FLY=[6,7,8,9,1,2,3,4]
PALACE_TRIGRAM={1:"坎",2:"坤",3:"震",4:"巽",5:"中",6:"乾",7:"兑",8:"艮",9:"离"}

def palace(idx):
    if idx<=1 or idx==23:return 1
    if 2<=idx<=4:return 8
    if 5<=idx<=7:return 3
    if 8<=idx<=10:return 4
    if 11<=idx<=13:return 9
    if 14<=idx<=16:return 2
    if 17<=idx<=19:return 7
    return 6

def fly(center,forward):
    out=[None,None,None,None,center,None,None,None,None]
    for i,p in enumerate(FLY):
        n=(center+i+1)%9 if forward else (center-i-1+9)%9
        out[p-1]=9 if n==0 else n
    return out

def forward(center, mountain_idx):
    if center==5:return MOUNTAINS[mountain_idx]["yin_yang"]=="阳"
    target=MOUNTAINS[mountain_idx]
    for m in MOUNTAINS:
        if m["trigram"]==PALACE_TRIGRAM[center] and m["yuan_long"]==target["yuan_long"]:
            return m["yin_yang"]=="阳"
    raise AssertionError((center,mountain_idx))

def situation(yun,sit,face,period,mountain,facing):
    sp=palace(sit);fp=palace(face)
    sit_m=mountain[sp-1]==yun;face_f=facing[fp-1]==yun
    sit_f=facing[sp-1]==yun;face_m=mountain[fp-1]==yun
    if sit_m and face_f:return "旺山旺向"
    if sit_m and sit_f:return "双星会坐"
    if face_m and face_f:return "双星会向"
    if face_m and sit_f:return "上山下水"
    return "未成四大局"

def start_year(yun):return 1864+(yun-1)*20

cases=[]
for yun in range(1,10):
    year=start_year(yun)+10
    period=fly(yun,True)
    for sit in range(12):
        face=(sit+12)%24
        shan_center=period[palace(sit)-1]
        xiang_center=period[palace(face)-1]
        mountain=fly(shan_center,forward(shan_center,sit))
        facing=fly(xiang_center,forward(xiang_center,face))
        cases.append({
            "id":f"yun{yun}-{MOUNTAINS[sit]['name']}-{MOUNTAINS[face]['name']}",
            "yun":yun,"year":year,"sit":sit,"face":face,
            "period_stars":period,"mountain_stars":mountain,"facing_stars":facing,
            "four_situation":situation(yun,sit,face,period,mountain,facing),
        })
doc={
 "version":1,
 "basis":"《沈氏玄空学》下卦正向：运盘顺飞，山向星按三元龙阴阳定顺逆。",
 "policy":{"plate":"lower_plate_forward_direction","sit_face":"opposite_24_mountain"},
 "scope":{"yun_count":9,"sit_face_pairs_per_yun":12,"case_count":len(cases)},
 "cases":cases,
}
out=ROOT/"engine/internal/engine/xuankong/testdata/flying_star_matrix_golden.json"
out.write_text(json.dumps(doc,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
print(f"wrote {len(cases)} cases")
