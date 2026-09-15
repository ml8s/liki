#!/usr/bin/env python3
"""Generate complete BaZhai ming-gua cycle fixture from the classical formula."""
from __future__ import annotations

import json

GAN_ZHI = {1:"坎",2:"坤",3:"震",4:"巽",6:"乾",7:"兑",8:"艮",9:"离"}
WEST = {"乾","兑","艮","坤"}

def ming_gua(gender: str, year: int) -> tuple[str,str]:
    y=year%100
    if year<2000:
        n=(100-y)%9 if gender=="male" else ((y-4)%9+9)%9
    else:
        n=(99-y)%9 if gender=="male" else (y+6)%9
    n=n or 9
    if n==5:
        n=2 if gender=="male" else 8
    name=GAN_ZHI[n]
    return name,("西四命" if name in WEST else "东四命")

cases=[]
for year in range(1900,2100):
    for gender in ("male","female"):
        gua,group=ming_gua(gender,year)
        cases.append({"gender":gender,"year":year,"gua":gua,"group":group})
doc={
 "version":1,
 "basis":"《八宅明镜》通行命卦公式；2000前后公式分段；5男寄坤、5女寄艮。",
 "policy":{"year_boundary":"gregorian_calendar_year","coverage":"1900-2099_all_gender"},
 "scope":{"case_count":len(cases)},
 "cases":cases,
}
open("engine/internal/engine/bazhai/testdata/minggua_cycle_golden.json","w",encoding="utf-8").write(json.dumps(doc,ensure_ascii=False,indent=2)+"\n")
print(f"wrote {len(cases)} cases")
