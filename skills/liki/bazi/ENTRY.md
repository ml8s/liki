# Liki 八字 — 八字 / 紫微
> 覆盖八字排盘、十神格局、大运流年，以及紫微命盘、宫位星曜、大限流年；两者通过引擎断语合参。

## Python 工具契约

bazi 域只通过 Python 工具层调用，不直接调用 RPC。

- 工具 schema：`bazi/tools/skill-tools.json`
- CLI：`python3 bazi/tools/agent_cli.py`
- Windows CLI：`bazi/tools/agent_cli.cmd`
- 输入：`{"fn":"<工具名>","args":{...}}`
- 输出：stdout JSON；`ok=true` 读 `data`，`ok=false` 读 `error`。
- CLI 启动时校验 engine 版本和必需 RPC；不满足即 fail closed。

## 路由

| 用户问题 | 入口 |
|---|---|
| 婚姻 / 感情 / 合婚 | `bazi/app/marriage.md`、`bazi/app/compatibility.md` |
| 事业 / 财运 / 学业 | `bazi/app/career.md`、`bazi/app/wealth.md`、`bazi/app/study.md` |
| 健康 / 性格 / 家庭 | `bazi/app/health.md`、`bazi/app/personality.md`、`bazi/app/family.md` |
| 排盘 / 随便看看 / 快速扫描 | `bazi/app/mingshu.md` |
| 完整命书 / 全盘报告 | `bazi/app/mingshu-full.md` |
| 占卜 / 风水 / 起名 | 返回根 `SKILL.md` 的领域路由 |
| 其他 | 确认意图后选择相近卡 |

## 领域工具

- 工具 schema：`bazi/tools/skill-tools.json`
- CLI：`python3 bazi/tools/agent_cli.py`，Windows 使用 `bazi/tools/agent_cli.cmd`

## 核心流程

| 步骤 | 条件 | 动作 | 产物 |
|---|---|---|---|
| 1 | 用户给具体时刻 | `city_coords` → `full_paipan(correct=true)` | 带 pan_digest 的完整 `pan` |
| 1 | 用户已明确时辰 | `full_paipan(correct=false)` | 带 pan_digest 的完整 `pan` |
| 1 | 时间距时辰交界 ≤30 分钟 | 先排盘，再提示相邻时辰校准 | 带校准提示的结果 |
| 1 | 时辰模糊或未知 | 收集候选与事件 → `calibrate`；读取 `bazi/domains/bazi/calibration.md`、`bazi/domains/ziwei/calibration.md` | 候选置信度 |
| 2 | 场景明确 | 读取对应 app 卡并执行 | 场景所需断言 |
| 3 | 本命分析 | `query(rule, pan)` | 八字 / 紫微 / 合参断语 |
| 4 | 应期分析 | `yearly_range(pan, start, end, rules)` | 含当前年来源与八字 / 紫微年界说明的流年事件与候选应期 |
| 5 | 双人关系 | `bond(pan_a, pan_b)` | 合盘因子 |
| 6 | 输出 | 按 app 模板综合 | 结论 + 依据链 |

同一会话复用 `full_paipan` 返回的完整 `pan`；多领域问题走主场景全流程，次领域仅查询佐证。信号冲突按 `bazi/domains/bazi/caijue.md` 裁决；有真实事件时取 3-5 段已发生时段验证，无验证则标注。用户只给出生地与时钟时间，时区、夏令时与真太阳时由工具链处理。

`query` 口径：`夫妻` / `官禄` / `财帛` 分别返回对应紫微宫位域；`十神` 返回八字结构及其跨人生领域信号；`用神` 除断语外返回 yong_shen_context 字段，直接包含 engine 三派、五行状态与十神状态。需要单一生活领域时优先使用宫位域，不把 `十神` 当作唯一领域查询。

## 硬边界

- `full_paipan` 返回的 `pan` 不可变；`query` / `yearly_range` / `bond` / `calibrate` 必须校验 `pan_digest`。
- `query` / `yearly_range` 只传入 `full_paipan` 的完整 `pan`；跨度含端点最多 120 年。
- `correct=true` 必须提供经度；缺出生地时先问城市，仍缺失则停止校正排盘。
- 出生时间距时辰交界 ≤30 分钟时，必须提示可校准；不得把该时辰表述为唯一确定事实。
- 考时证据不足时输出「考时证据不足」；用户指定候选时标注「未经考时确认」。
- 八字与紫微单侧结论各自输出；合参结论仅来自 common 断言。

## 输出契约

- 涉及生育、健康、年龄窗口、重大财务或事业变动时，附现实专业确认提示；不得把应期写成必然结果。

## 异常处理

| 异常 | 处理 |
|---|---|
| `RPCError` / `ValueError` | 网络异常可重试；参数错误修正后重试 |
| 城市未收录 / 某年 `error` | 请用户给附近较大城市；或输出该年数据缺失 |
