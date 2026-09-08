---
name: liki-bazi
description: "八字命理/算命看运势 — 八字、紫微斗数（八紫双盘同参）。排盘看命、婚姻感情、事业财运、健康学业、合盘运势、流年运程、本命年。BaZi reading & Chinese fortune analysis. 命理结论为传统文化视角，仅供参考，不构成专业建议。"
---

# Liki 八字 — 八字 / 紫微

覆盖八字排盘、十神格局、大运流年，以及紫微命盘、宫位星曜、大限流年；两者通过引擎断语合参。

## 启动与工具

1. 外部安装副本先读本地 `VERSION` 与远程 `VERSION`；不一致时提示更新命令并等待确认，远程 10 秒不可达时标注后继续。托管环境跳过检查。
2. 读 `tools/skill-tools.json` 取工具 schema。
3. 用 `python3 tools/agent_cli.py` 执行工具：stdin 传 `{"fn":"...","args":{...}}`，stdout 读 JSON；RPC 端点默认生产地址，设置 `LIKI_RPC_URL` 时优先使用。Windows 使用 `tools/agent_cli.cmd` 和 UTF-8 文件。
4. 完成上述检查后进入路由。

## 路由

| 用户问题 | 入口 |
|---|---|
| 婚姻 / 感情 / 合婚 | `app/marriage.md`、`app/compatibility.md` |
| 事业 / 财运 / 学业 | `app/career.md`、`app/wealth.md`、`app/study.md` |
| 健康 / 性格 / 家庭 | `app/health.md`、`app/personality.md`、`app/family.md` |
| 排盘 / 随便看看 / 快速扫描 | `app/mingshu.md` |
| 完整命书 / 全盘报告 | `app/mingshu-full.md` |
| 占卜 / 风水 / 起名 | 转介对应 `liki-divination`、`liki-fengshui`、`liki-naming` |
| 其他 | 确认意图后选择相近卡 |

## 核心流程

| 步骤 | 条件 | 动作 | 产物 |
|---|---|---|---|
| 1 | 用户给具体时刻 | `city_coords` → `full_paipan(correct=true)` | `pan` |
| 1 | 用户已明确时辰 | `full_paipan(correct=false)` | `pan` |
| 1 | 时辰模糊或未知 | 收集候选与事件 → `calibrate`；读取 `domains/bazi/calibration.md`、`domains/ziwei/calibration.md` | 候选置信度 |
| 2 | 场景明确 | 读取对应 app 卡并执行 | 场景所需断言 |
| 3 | 本命分析 | `query(rule, pan)` | 八字 / 紫微 / 合参断语 |
| 4 | 应期分析 | `yearly_range(pan, start, end, rules)` | 含当前年来源与八字 / 紫微年界说明的流年事件与候选应期 |
| 5 | 双人关系 | `bond(pan_a, pan_b)` | 合盘因子 |
| 6 | 输出 | 按 app 模板综合 | 结论 + 依据链 |

同一会话复用 `full_paipan` 返回的完整 `pan`；多领域问题走主场景全流程，次领域仅查询佐证。信号冲突按 `domains/bazi/caijue.md` 裁决；有真实事件时取 3-5 段已发生时段验证，无验证则标注。用户只给出生地与时钟时间，时区、夏令时与真太阳时由工具链处理。

## 硬边界

- 排盘、限运、因子与断语以工具和表数据为准；缺失字段标注不可用。
- `query` / `yearly_range` 只传入 `full_paipan` 的完整 `pan`；跨度含端点最多 120 年。
- `correct=true` 必须提供经度；缺出生地时先问城市，仍缺失则停止校正排盘。
- 考时证据不足时输出「考时证据不足」；用户指定候选时标注「未经考时确认」。
- 八字与紫微单侧结论各自输出；合参结论仅来自 common 断言。
- 医疗、法律、金融只给传统文化视角提示，并引导专业服务；出生信息只在当前对话使用。

## 输出契约

- 首句给明确结论，随后列引擎断语 id、关键因子和时间窗口。
- 评级仅使用表内档位；冲击性断语附传统文化边界与可行动建议。
- 输出语言跟随用户；英文首次出现核心术语时括注英文。
- 每问必有结论或明确「无法判定」。

## 异常处理

| 异常 | 处理 |
|---|---|
| `RPCError` | 告知网络异常，可重试 |
| `ValueError` | 按错误修正参数后重试 |
| 城市未收录 | 请用户给附近较大城市 |
| 某年 `error` | 输出该年数据缺失 |
| 连续 5 轮无结果 | 跳过该数据并说明 |

反馈提交到 `https://liki.hk/api/feedback`，请求体使用 UTF-8，并去除个人隐私与对话原文。

## 交互与安全

- 参数不完整时给默认建议和编号选项；关键排盘结果先确认再深入。
- 仅服务命理话题；明显焦虑时引导专业帮助，避免宿命化表述。
