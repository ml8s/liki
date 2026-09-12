---
name: qimen-yingqi
description: 奇门问事应期候选解释规则
---

# 奇门应期候选

> 命理依据：马星逢冲主动，空亡逢值或逢冲为填实候选；值符 / 值使落宫地支逢值或逢冲为引动候选。`qimen.chart` 输出地支 / 宫位事实，并标记与日干、时干、主柱、值符、值使、年命干与所选用神的相关性；LLM 不得解释 `related_to` 为空的候选。

## 引擎输出

`ying_qi`：

- `candidates[].type`：ma_xing、kong_wang、duty_star 或 duty_door
- `candidates[].mechanism`：候选机制（冲 / 值冲）
- `candidates[].branch`：地支事实
- `candidates[].gong`：宫位投影
- `candidates[].related_to[]`：相关符号列表，每项含 `symbol` 与 `role`
  - `role=pillar`：日干 / 时干 / 年命干
  - `role=lead`：当前方法主柱（hour=时柱；quarter 由 `lead_pillar_mode` 区分刻柱/时柱；day=日柱，month=月柱，year=年柱）
  - `role=yong_shen`：调用方传入的事象用神
  - `role=duty`：值符星或值使门落宫
- `candidates[].dates[]`：引擎给出的具体日期窗口，每项含 `date`、`branch`、`match`、`reason`
- `summary`：候选口径，不是结论

## 解读规则

| 候选 | 条件 | 解释 |
|------|------|------|
| 马星 | `type=ma_xing` 且 `related_to` 非空 | 马星逢冲为动期候选，主快、主动 |
| 空亡 | `type=kong_wang` 且 `related_to` 非空 | 逢值填实或逢冲为填实候选，主由虚转实 |
| 值符宫 | `type=duty_star` 且 `related_to` 非空 | 值符宫支逢值或逢冲为引动候选 |
| 值使宫 | `type=duty_door` 且 `related_to` 非空 | 值使宫支逢值或逢冲为引动候选 |
| 无关候选 | `related_to` 为空数组 | 只可列为盘面候选，不得进入本问应期 |

## 输出要求

1. 只解释 `related_to` 非空的候选。
2. 应期必须引用 `candidates[].dates[]` 或其他引擎时间事实，不得只写“近期”。
3. 日期窗口由引擎计算，LLM 不得补算日期。
4. 应期不改变吉凶；先断方向/成败，再给时间窗口。
5. 若所有候选均无关，明确说“本盘无与日干、时干、主柱、值符、值使、年命干或所选用神直接相关的引擎应期候选”。
