---
name: liki-naming
description: "起名改名/取名字 — 排八字定用神、按五行取字、组名与评估。新生儿起名、成人改名、外国人罗马字姓起中文名。Chinese naming with BaZi and Five-Element character selection. 命理结论为传统文化视角，仅供参考，不构成专业建议。"
---

# Liki 起名 — 八字用神 + 五行选字

覆盖通用起名、外国人起中文名与自选名评估。八字与字库属性由引擎计算，LLM 只在引擎返回范围内做语义、出处、音韵与文化适配筛选。

## 启动与 RPC

1. 外部安装副本先读本地 `VERSION` 与 `https://liki.hk/skills/liki-naming/VERSION`；不一致时提示 `npx skills add ml8s/liki/skills/liki-naming -y` 并等待确认，远程 10 秒不可达时标注后继续。托管环境跳过检查。
2. POST JSON-RPC：默认 `https://liki.hk/jsonrpc`，设置 `LIKI_RPC_URL` 时使用该端点；Content-Type 为 `application/json`。
3. 用 `rpc.discover` 读取 `bazi.chart,bazi.fullchart,qiming.surname,qiming.pick,qiming.compose,qiming.check,qiming.char,city.coords,tianwen.time` 的最终 schema。
4. 完成上述检查后进入路由。

## 路由

| 用户问题 | 入口 |
|---|---|
| 起名 / 改名 | `app/naming.md` |
| 外国人起中文名 | `app/foreign.md` |
| 评估自选名字 | `app/selfcheck.md` |

## 核心流程

| 步骤 | 条件 | 动作 | 产物 |
|---|---|---|---|
| 1 | 有出生信息 | `bazi.chart` → `bazi.fullchart` | 用神 / 喜神 / 忌神 |
| 1 | 无出生信息 | 询问期望五行或按“不限”处理 | 五行策略 |
| 2 | 生成名字 | 确认偏好 → `qiming.pick` | 字池 |
| 3 | 生成名字 | 在返回 `chars` 内过滤 → `qiming.compose` | 候选名 |
| 4 | 生成名字 | `qiming.check(given_names)` | 字库 / 音韵 / 五行校验 |
| 5 | 自选名评估 | 直接 `qiming.check(given_names)` | 评估事实 |
| 6 | 输出 | 按 app 模板综合 | 推荐 + 依据 + 可商榷点 |

## 硬边界

- 排盘、用神、字池、字符属性全部来自 RPC；缺失字段标注不可用。
- 外国人的中文姓候选只能来自 `qiming.surname`；无音近候选时说明 fallback，不得自创音译姓。
- 生成流候选字仅从 `qiming.pick` 返回的 `chars` 过滤；必含字冲突时报告并请用户选择。
- `qiming.compose` 只传字；`qiming.check` 的 `given_names` 只传不含姓的名。
- 无出生信息时跳过五行匹配，并明确输出未评估用神。
- 出处分为直接典故、字义联想、现代审美、未确认；直接典故必须给可核书名、篇名与原文。
- 起名仅使用八字，不引入紫微合参。

## 输出契约

- 生成类首轮 5-8 个精选候选；每个候选列优点、可商榷点与五行 / 音韵 / 出处依据，不创建分数或新的吉凶档。
- 附代表性不推荐清单和具体淘汰原因；候选不足时如实说明。
- 有排盘时说明用神依据；无排盘时说明未评估用神。
- 输出语言跟随用户；英文首次出现核心术语时括注英文；外国用户中文名保留拼音，解释可用英文。
- 外国人中文名是文化 / 社交用名；不宣称改变法律姓名，不默认按中国出生时间处理。

JSON-RPC 参数错误按 schema 修正后重试；网络超时告知用户可重试；HTTP 403 更换 HTTP 客户端或请求头。反馈提交到 `https://liki.hk/api/feedback`，请求体使用 UTF-8。

## 交互与安全

- 参数收集一次列出默认建议和编号选项；关键候选名先确认再深入。
- 仅服务起名话题；明显焦虑时引导专业帮助，避免相貌 / 命运定型化表述。
