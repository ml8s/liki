#!/usr/bin/env python3
"""Generate two-oracle calendar consensus cases for HuangLi QueryDate."""
from __future__ import annotations

import datetime as dt
import json
from pathlib import Path
import sys

from lunar_python import Solar
import sxtwl

ROOT=Path(__file__).resolve().parents[3]
CORE=json.loads((ROOT/"tests/fixtures/domain_oracle/huangli_core.json").read_text())
SEQ=CORE["jianchu_sequence"]
QING_START=CORE["qing_long_start_all_12"]
EVENTS=CORE["event_rules"]
ZHI="子丑寅卯辰巳午未申酉戌亥"
GAN="甲乙丙丁戊己庚辛壬癸"
STARS=["青龙","明堂","天刑","朱雀","金匮","天德","白虎","玉堂","天牢","玄武","司命","勾陈"]
PATHS=["黄道","黄道","黑道","黑道","黄道","黄道","黑道","黄道","黑道","黑道","黄道","黑道"]
EVENT_NAMES=list(EVENTS)

def consensus(date):
    lp=Solar.fromYmd(date.year,date.month,date.day).getLunar()
    sx=sxtwl.fromSolar(date.year,date.month,date.day)
    day=(lp.getDayGan(),lp.getDayZhi())
    sxday=(GAN[sx.getDayGZ().tg],ZHI[sx.getDayGZ().dz])
    yue=lp.getMonthZhiExact(); sxyue=ZHI[sx.getMonthGZ().dz]
    if day!=sxday or yue!=sxyue:
        return None
    return day,yue

def make_case(date, event):
    values=consensus(date)
    if values is None:
        return None
    (day_gan,day_zhi),yue_zhi=values
    jian=SEQ[(ZHI.index(day_zhi)-ZHI.index(yue_zhi))%12]
    offset=(ZHI.index(day_zhi)-ZHI.index(QING_START[yue_zhi]))%12
    rule=EVENTS[event]
    suitability="recommended" if jian in rule["suitable"] else ("unsuitable" if jian in rule["forbidden"] else "possible")
    return {
      "date":date.isoformat(),"event":event,
      "expected":{
        "day_gan":day_gan,"day_zhi":day_zhi,"yue_zhi":yue_zhi,"jian_chu":jian,
        "huangdao_name":STARS[offset],"huangdao_path":PATHS[offset],"suitability":suitability
      },
      "consensus":["lunar-python","sxtwl"],
    }

cases=[]
date=dt.date(2024,1,1)
while date<=dt.date(2027,12,31):
    event=EVENT_NAMES[len(cases)%len(EVENT_NAMES)]
    case=make_case(date,event)
    if case: cases.append(case)
    date += dt.timedelta(days=16)

doc={
 "version":1,
 "basis":"《协纪辨方书》建除、青龙黄道起例；日柱与节月由 lunar-python/sxtwl 共识。",
 "policy":{"day_boundary":"local_midnight_utc8","consensus":"all_oracles_must_agree"},
 "scope":{"case_count":len(cases),"events":len(EVENT_NAMES),"jianchu_states":12},
 "sources":[{"id":"lunar-python","version":"1.4.8"},{"id":"sxtwl","version":"2.0.7"}],
 "known_limits":["dates with date-level solar-term disagreement between oracles are skipped"],
 "cases":cases,
}
out=ROOT/"engine/internal/engine/huangli/testdata/query_date_matrix_golden.json"
out.write_text(json.dumps(doc,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
print(f"wrote {len(cases)} cases")
