---
name: app-qimen
description: 奇门问卦 — snapshot 生成、结构化 answer 与追问
依赖域: qimen
---

# 奇门问卦

## 依赖的领域知识

[必读] `domains/qimen/yongshen.md` + `bamen.md` + `jiuxing.md` + `bashen.md` + `yingqi.md` + `quarter.md`
[专占] `domains/qimen/lost-property.md` + `thief-capture.md` + `missing-person.md` + `capture-escape.md`

## 流程

| 步骤 | 动作 | 产物 |
|---|---|---|
| 1 | 确认单一目标、地点和事项 | 可起局的问题 |
| 2 | 调 `qimen_snapshot`：传 `question`、`city` 或 `longitude`，普通问事传 `matter`，高级用户可直接传 `yong_shen`，专占传 `rule` | immutable snapshot |
| 3 | 调 `qimen_ask(snapshot, message)` | 结构化 answer |
| 4 | 解释 factors 中的用神落宫、门星神、生克、空亡、马星和应期 | 方向、态势与时机 |
| 5 | 追问继续传同一 snapshot | 不重排上下文 |

## 常用参数

用户没有明确要求，不传方法参数；默认 `scope=hour` / `school=zhuanpan` / `dingju_method=chaibu`，不传其他方法参数。不能从“更细”“传统”“准确”推断高级方法。

| 用户问题 | 参数 |
|---|---|| 钥匙丢了，能不能找到 | `rule=lost_property` |
| 东西被偷 / 逃走的人 / 偷者画像 | `qimen_snapshot(rule=thief_capture/capture_escape/thief_profile)` |
| 工作能不能升、财能不能求、婚姻如何 | `matter=career/wealth/relationship` |
| 家人走失 | `qimen_snapshot(matter=missing_person, rule=missing_person)` |
| 看今天整体 | `scope=day` |
| 用置闰盘看现在 | `dingju_method=zhirun` |
| 用洛书飞盘看这件事 | `school=luoshu_feipan` |
| 用十分钟刻家看看 | `scope=quarter` + `quarter_rule=ten_minute_sanyuan` |
| 用十二分钟十分局看看 | `scope=quarter` + `quarter_rule=twelve_minute_ten_division` |
| 用金函玉镜看今天 | `scope=day` + `school=jinhan_yujing` |
| 刻家 / 流派 / 定局 | 按用户明确要求传对应高级参数 |

## 边界

- 缺地点时先要城市或经度；缺时间默认用服务端当前时间。
- `matter` 与 `yong_shen` 互斥。
- 金函玉镜不输出局数、用神、值符值使与应期。
- 奇门偏当前决策与近势；长期命局转八字。

## 输出模板

```text
结论：一句话方向 / 时机 / 决策倾向。
用神：符号、落宫、旺衰、空亡 / 马星。
态势：支持与阻碍来源。
时机：answer 引用的应期候选。
建议：一个现实动作或核查条件。
边界：传统奇门视角，不承诺现实结果。
```
