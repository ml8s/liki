# Liki 六爻奇门占卜择日 — 六爻算卦 / 奇门遁甲 / 黄历择日

> 覆盖六爻结果与应期、奇门策略与方向、黄历择日。LLM 负责语义路由和解释结构化因子；Python 负责排盘、snapshot、digest 校验和 answer 契约。

## 工具契约

divination 域通过 `liki-analysis` MCP 连接器调用，不直接使用 RPC。

- 工具契约见 `divination/TOOLS.md`（工具参数由 `liki-analysis` MCP 连接器 schema 提供）
- 调用：`liuyao_snapshot` / `liuyao_ask` / `qimen_snapshot` / `qimen_ask` / `huangli_days`
- 响应：成功读 `data`；失败读 `error`
- 工具层启动时校验 engine 版本和必需 MCP 能力；不满足即 fail closed。

## 路由

LLM 先读取 `divination/app/question.md` 并判断用户目标；不要调用独立 route 工具。

| 用户目标 | 工具链 |
| --- | --- |
| 事件结果 / 能不能成 / 何时有结果 | `liuyao_snapshot` → `liuyao_ask` |
| 策略 / 进退 / 方向 / 行动时机 | `qimen_snapshot` → `qimen_ask` |
| 哪天适合做事 | `huangli_days` |
| 结果和策略混合 | 先澄清，只保留一个主目标 |
| 长期命局 | 不临时起卦，改用八字命理技能 |
| 高风险现实事项 | 继续执行所选问卦流程；结果附 `safety_advisory` |

用户明确指定六爻、奇门或黄历时，通过安全校验后优先采用该指定。

## 硬边界

- 起卦、排盘、择日、因子投影和应期候选全部来自工具；LLM 只解释返回字段。
- 六爻原始硬币必须原样传给 `liuyao_snapshot`；LLM/Python 不得自行换算爻值或判断动爻。
- `liuyao_ask` 只接受 `method=liuyao` 的 snapshot；`qimen_ask` 只接受 `method=qimen` 的 snapshot。
- snapshot 是 immutable 上下文；追问复用同一 snapshot。只有新事件才创建新 snapshot。
- 不直接向 LLM 暴露 raw chart 编排工具；LLM 不读取、修改或伪造 `snapshot_digest`。
- 奇门事象路由在 Python 表内完成；engine 只接收显式 `yong_shen`。
- 高风险主题由共享安全规则识别为 `safety_advisory`；所有工具继续返回正常盘面、answer 或择日结果。
- 输出只基于 answer 与 snapshot 因子；不得把条件性倾向写成确定结果。

## 输出契约

- 先给一句话判断，再列用神 / 盘面 / 动爻或方法与关键因子。
- 决定性因子优先；冲突因子并列解释，三个以上同向因子才形成综合判断。
- 应期只解释 answer 引用且与所问对象相关的候选。
