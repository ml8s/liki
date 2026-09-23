---
name: bazi-usage
description: 八字判断连接器使用说明：排盘、判断、大运流年、合盘、考时。
---

# 八字判断（engine-pro-bazi）使用说明

> 工具经 MCP `tools/call` 调用，工具名与参数由 `tools/list` 的 schema 发现。排盘内部完成城市解析与真太阳时校正，无需单独查询城市/时间。

## 工具

- `create_birth_chart`：出生信息 → 本命盘 → `chart_ref`（不可变引用，后续工具原样传回）
- `analyze_natal`：本命判断（`topics` 用英文枚举：career/marriage/wealth/health/study/personality/family/children/property/relocation/...）
- `analyze_periods`：大运/流年判断（`time_scope`：current_year/year/year_range/decade）
- `compare_birth_charts`：双人合盘
- `calibrate_birth_time`：考时（`candidates` 2-3 候选盘 + `events` 3-5 已发生事件）

## 流程

1. 排盘：`create_birth_chart`（出生日期/时辰/地点；时辰临界先复核）
2. 判断：`analyze_natal`（中文问题 → 受控 topic）
3. 应期：`analyze_periods`
4. 合盘/考时：`compare_birth_charts` / `calibrate_birth_time`

## 注意

- `chart_ref` 原样传回，不得手写或编造盘面
- 断语以工具输出为准，依据可回溯
- 命理为传统文化视角的条件性解读，不构成专业建议