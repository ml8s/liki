---
name: qimen-quarter
description: 刻家奇门的方法边界
---

# 刻家奇门

| 口径 | 参数 | 规则 |
|---|---|---|
| 十分钟三元刻家 | `quarter_rule=ten_minute_sanyuan`（默认） | 一时辰 12 刻，每刻 10 分钟；刻柱主盘 |
| 十二分钟十分局刻家 | `quarter_rule=twelve_minute_ten_division` | 一时辰 10 刻，每刻 12 分钟；初局取时家局，阳顺阴逆；时柱主盘 |

- 仅支持 `scope=quarter` + `school=zhuanpan`。
- `method.lead_pillar` 是唯一主柱干支事实源；`method.quarter.lead_pillar_mode` 说明 `quarter=刻柱` 或 `hour=时柱`。
- 十二分钟十分局可用 `base_dingju_method` 指定基准时家定局法，可选拆补 / 置闰 / 茅山，默认拆补。
- 十二分钟十分局可用 `dun_source` 选择时支分遁或节气分遁，默认时支分遁；用 `hour_boundary` 选择晚子时或子正起子时，默认晚子时。
- 局数、刻序与主柱口径由引擎输出；LLM 不得重算或改写。
- 十二分钟十分局是具名时间切分规则，不冒充通用传统标准。
