---
name: app-qimen
description: 奇门问事 — 方向、时机、用神落宫与决策
依赖域: qimen
---

# 奇门问事
## 依赖的领域知识

[必读] - qimen: `domains/qimen/yongshen.md`（用神）+ `bamen.md`（八门）+ `jiuxing.md`（九星）+ `bashen.md`（八神 / 洛书飞盘九神）+ `yingqi.md`（应期）；金函玉镜另读 `jinhan.md`

## 📖 流程

| 步骤 | 条件 / 目标 | 动作 | 产物 |
|---|---|---|---|
| 1 | 场景复核 | 已由 `divination_route` 选中 `qimen` 时才进入本卡 | 奇门适用性确认 |
| 2 | 一次读盘 | `qimen_read`：传 `question`、`city` 或 `longitude`、可选 `time`；普通问事传 `matter`，高级用户可直接传 `yong_shen`；`matter` 与 `yong_shen` 互斥，专占传 `rule` | 时间、盘面、snapshot、专占候选 |
| 3 | 专占解释 | `rule` 命中时分别读 `domains/qimen/lost-property.md`、`thief-capture.md`、`missing-person.md`、`capture-escape.md`；候选并列，不临场排序 | 表内断语候选 |
| 4 | 标准盘解读 | 门、星、神、生克、星宫五行、空亡、马星、五不遇时、格局、十干克应、门迫门制；刻家按 `lead_pillar_mode` 解释主柱 | 方向 / 时机 / 阻碍 |
| 5 | 应期 | 读取 `domains/qimen/yingqi.md`，只解释 `related_to` 非空候选 | 时间窗口 |
| 6 | 金函玉镜解读 | 按 `jinhan.md` 只解释金函九星与八门；十二神仅展示，不下断语 | 方向 / 阻碍 |
| 7 | 报告骨架 | 调用 `qimen_report(action=template, read_result=...)` | 方法、snapshot、专占与应期引用 |
| 8 | 报告审计 / 会话 | `qimen_report(action=validate, ...)` 通过后，用 `qimen_session(action=create, ...)` 固化原局与首次结论 | 审计结果 + 可追问 session |
| 9 | 输出 | 按模板综合 | 决策 + 依据 + 建议 |
## 用户说法路由
用户没有明确要求，不传方法参数；默认 `scope=hour` / `school=zhuanpan` / `dingju_method=chaibu`。不能从“更细”“传统”“准确”推断高级方法。

| 用户说法 | 参数 |
|---|---|
| `钥匙丢了，能不能找到` | `qimen_read(rule=lost_property)` |
| `东西被偷了…` / `逃走的人…` / `偷者画像` | `qimen_read(rule=thief_capture/capture_escape/thief_profile)` |
| `工作能不能升` / `这笔财能不能求` / `婚姻如何` | `matter=career/wealth/relationship` |
| `家人走失了` | `qimen_read(matter=missing_person, rule=missing_person)` |
| `用置闰盘看现在` | `dingju_method=zhirun` |
| `用洛书飞盘看这件事` | `school=luoshu_feipan` |
| `看今天整体` | `scope=day` |
| `用十分钟刻家看看` | `scope=quarter` + `quarter_rule=ten_minute_sanyuan` |
| `用十二分钟十分局看看` | `scope=quarter` + `quarter_rule=twelve_minute_ten_division` |
| `用金函玉镜看今天` | `scope=day` + `school=jinhan_yujing` |

## 边界条件

| 异常场景 | 处理方式 |
|---|---|
| 用户只给钟表时间 | `qimen_read` 内部取地点经纬度并换算真太阳时 |
| 当前事 / 更细时辰事 / 今日 / 本月 / 今年 | 分别传 `scope=hour/quarter/day/month/year` |
| 用户要求洛书飞盘 / 转盘 / 鸣法飞盘 | 分别传 `school=luoshu_feipan/zhuanpan/mingfa_feipan` |
| 用户要求置闰 / 拆补 / 茅山 | 置闰与拆补仅 hour/day 传 `dingju_method=zhirun/chaibu`；茅山仅 hour + 转盘传 `dingju_method=maoshan`；month/year 不传 |
| 用户要求十分钟刻家 | 传 `scope=quarter` + `school=zhuanpan` + `quarter_rule=ten_minute_sanyuan`；按 `quarter.md` 解释 |
| 用户要求十二分钟十分局 | 传 `scope=quarter` + `school=zhuanpan` + `quarter_rule=twelve_minute_ten_division`；基准定局法用 `base_dingju_method`，分遁用 `dun_source`，时辰界用 `hour_boundary`；按 `quarter.md` 解释 |
| 用户要求金函玉镜日家 | 传 `scope=day` + `school=jinhan_yujing`，不传 `dingju_method/quarter_rule/yong_shen/birth_date`，按 `jinhan.md` 解释 |
| 阴盘 / 非鸣法九门飞盘 / schema 未暴露方法 | 说明未支持并改用可用方法或停止；不近似替代。鸣法飞盘仅 hour + 拆补，茅山法仅 hour + 转盘 |
| 问题偏中长期 | 说明奇门偏当前决策与近势，可配合六爻 |

## 📖 输出模板

| 输出 | 内容 |
|---|---|
| 方法 | `scope_name`、`school_name`、`dingju_method_name` / `quarter_rule_name`、主柱、刻序（如请求 quarter）与 `zhi_run_state` |
| 用神 | matter.name、符号、落宫、天盘干、空亡 / 马星 |
| 盘面 | 门星神组合、生克、值符值使 |
| 金函玉镜 | 二至阴阳、日柱、逐宫星门；不输出局数、用神、值符值使与应期 |
| 结论 / 应期 | 一句话方向 / 时机 / 成败；相关应期候选与 `dates[]` 时间窗口 |
| 报告审计 | report v1 引用核对；未知证据、方法不匹配或禁语会被拒绝 |
| 追问 | session 固化原局时间、地点、方法、snapshot 与首次结论 |
